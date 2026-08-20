import { useEffect, useRef, useState } from 'react';

const EMPTY_STATS = {
  fps: 0,
  bitrateMbps: 0,
  rttMs: 0,
  packetLoss: 0,
  jitterMs: 0,
  width: 0,
  height: 0,
  direction: 'receive',
  ready: false,
};

function number(value, fallback = 0) {
  return typeof value === 'number' && Number.isFinite(value) ? value : fallback;
}

function findSelectedCandidatePair(report) {
  let selectedPair = null;

  report.forEach((item) => {
    if (item.type === 'candidate-pair' && item.state === 'succeeded' && item.nominated) {
      selectedPair = item;
    }
  });

  if (selectedPair) return selectedPair;

  report.forEach((item) => {
    if (item.type !== 'transport' || !item.selectedCandidatePairId) return;
    const pair = report.get(item.selectedCandidatePairId);
    if (pair) selectedPair = pair;
  });

  return selectedPair;
}

function pickVideoRtp(report, direction) {
  let selected = null;

  report.forEach((item) => {
    if (item.type !== direction) return;
    if (item.kind !== 'video' && item.mediaType !== 'video') return;
    if (item.isRemote) return;

    // При simulcast может быть несколько outbound RTP записей. Для оверлея
    // берём активную запись с наибольшим количеством переданных/полученных байт.
    if (!selected || number(item.bytesSent ?? item.bytesReceived) > number(selected.bytesSent ?? selected.bytesReceived)) {
      selected = item;
    }
  });

  return selected;
}

/**
 * Реальные WebRTC-метрики активной демонстрации экрана.
 * Для удалённой демки показывает фактически получаемый поток у зрителя,
 * для собственной — фактически отправляемый поток.
 */
export default function useScreenShareStats(track) {
  const [stats, setStats] = useState(EMPTY_STATS);
  const previousRef = useRef({ bytes: 0, timestamp: 0, packets: 0, lost: 0 });

  useEffect(() => {
    if (!track || typeof track.getRTCStatsReport !== 'function') {
      setStats(EMPTY_STATS);
      previousRef.current = { bytes: 0, timestamp: 0, packets: 0, lost: 0 };
      return undefined;
    }

    let disposed = false;
    const isLocal = Boolean(track.isLocal);
    const direction = isLocal ? 'outbound-rtp' : 'inbound-rtp';

    previousRef.current = { bytes: 0, timestamp: 0, packets: 0, lost: 0 };

    const sample = async () => {
      try {
        const report = await track.getRTCStatsReport();
        if (!report || disposed) return;

        const rtp = pickVideoRtp(report, direction);
        if (!rtp) return;

        const timestamp = number(rtp.timestamp, performance.now());
        const bytes = number(isLocal ? rtp.bytesSent : rtp.bytesReceived);
        const packets = number(isLocal ? rtp.packetsSent : rtp.packetsReceived);
        const lost = number(rtp.packetsLost);
        const previous = previousRef.current;

        let bitrateMbps = 0;
        if (previous.timestamp > 0 && timestamp > previous.timestamp && bytes >= previous.bytes) {
          const seconds = (timestamp - previous.timestamp) / 1000;
          bitrateMbps = ((bytes - previous.bytes) * 8) / seconds / 1_000_000;
        } else if (number(track.currentBitrate) > 0) {
          bitrateMbps = number(track.currentBitrate) / 1_000_000;
        }

        let packetLoss = 0;
        if (!isLocal && previous.timestamp > 0) {
          const packetDelta = packets - previous.packets;
          const lostDelta = lost - previous.lost;
          const total = packetDelta + Math.max(0, lostDelta);
          if (total > 0) packetLoss = (Math.max(0, lostDelta) / total) * 100;
        }

        let rttMs = 0;
        let jitterMs = number(rtp.jitter) * 1000;

        // Для outbound статистики feedback от получателя обычно лежит в
        // remote-inbound-rtp. Для обоих направлений дополнительно берём RTT
        // выбранной ICE candidate pair, если браузер его предоставляет.
        if (isLocal) {
          report.forEach((item) => {
            if (
              item.type === 'remote-inbound-rtp' &&
              (item.kind === 'video' || item.mediaType === 'video')
            ) {
              if (number(item.roundTripTime) > 0) rttMs = number(item.roundTripTime) * 1000;
              if (number(item.jitter) > 0) jitterMs = number(item.jitter) * 1000;
              if (typeof item.fractionLost === 'number') packetLoss = Math.max(0, item.fractionLost * 100);
            }
          });
        }

        const candidatePair = findSelectedCandidatePair(report);
        if (candidatePair && number(candidatePair.currentRoundTripTime) > 0) {
          rttMs = number(candidatePair.currentRoundTripTime) * 1000;
        }

        previousRef.current = { bytes, timestamp, packets, lost };

        if (disposed) return;

        setStats({
          fps: number(rtp.framesPerSecond),
          bitrateMbps: Math.max(0, bitrateMbps),
          rttMs: Math.max(0, rttMs),
          packetLoss: Math.max(0, packetLoss),
          jitterMs: Math.max(0, jitterMs),
          width: number(rtp.frameWidth),
          height: number(rtp.frameHeight),
          direction: isLocal ? 'send' : 'receive',
          ready: true,
        });
      } catch (error) {
        if (!disposed) {
          console.debug('Не удалось получить WebRTC-статистику демонстрации:', error);
        }
      }
    };

    sample();
    const timer = window.setInterval(sample, 1000);

    return () => {
      disposed = true;
      window.clearInterval(timer);
    };
  }, [track]);

  return stats;
}
