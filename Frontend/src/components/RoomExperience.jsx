import {
  LiveKitRoom,
  RoomAudioRenderer,
  TrackToggle,
  VideoTrack,
  useLocalParticipant,
  useParticipants,
  useRoomContext,
  useTracks,
} from '@livekit/components-react';
import { RoomEvent, Track } from 'livekit-client';
import {
  LogOut,
  Maximize2,
  Mic,
  MicOff,
  Minimize2,
  MonitorPlay,
  MonitorUp,
  MonitorX,
  MoreVertical,
  Users,
  Volume1,
  Volume2,
  VolumeX,
  X,
} from 'lucide-react';
import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import { createPortal } from 'react-dom';
import useScreenShareStats from '../hooks/useScreenShareStats';
import Avatar from './Avatar';

const VOLUME_MAX = 200;

function trackKey(trackRef) {
  return (
    trackRef.publication?.trackSid ||
    `${trackRef.participant.identity}-screen`
  );
}

/* ===========================
   Звук входа/выхода участника
=========================== */

let sharedAudioContext = null;

function getAudioContext() {
  if (typeof window === 'undefined') return null;
  const AudioContextClass = window.AudioContext || window.webkitAudioContext;
  if (!AudioContextClass) return null;

  if (!sharedAudioContext) {
    sharedAudioContext = new AudioContextClass();
  }

  if (sharedAudioContext.state === 'suspended') {
    sharedAudioContext.resume().catch(() => {});
  }

  return sharedAudioContext;
}

// Короткий синтезированный сигнал (без внешних аудиофайлов): восходящие
// две ноты — кто-то вошёл, нисходящие — кто-то вышел.
function playPresenceChime(kind) {
  const ctx = getAudioContext();
  if (!ctx) return;

  try {
    const now = ctx.currentTime;
    const notes = kind === 'join' ? [523.25, 659.25] : [659.25, 493.88];
    const noteDuration = 0.13;

    notes.forEach((freq, i) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.type = 'sine';
      osc.frequency.value = freq;

      const start = now + i * noteDuration;
      gain.gain.setValueAtTime(0.0001, start);
      gain.gain.linearRampToValueAtTime(0.16, start + 0.015);
      gain.gain.exponentialRampToValueAtTime(0.0001, start + noteDuration);

      osc.connect(gain).connect(ctx.destination);
      osc.start(start);
      osc.stop(start + noteDuration + 0.02);
    });
  } catch (error) {
    console.error('Не удалось воспроизвести звук уведомления:', error);
  }
}

function VolumeIcon({ volume, size = 14 }) {
  if (volume === 0) return <VolumeX size={size} />;
  if (volume < 100) return <Volume1 size={size} />;
  return <Volume2 size={size} />;
}

/**
 * Закрывает всплывающее меню по клику вне его, скроллу или ресайзу окна.
 */
function useOutsideDismiss(open, onClose, refs) {
  useEffect(() => {
    if (!open) return undefined;

    const handlePointer = (event) => {
      const insideAny = refs.some((ref) => ref.current?.contains(event.target));
      if (!insideAny) onClose();
    };

    document.addEventListener('mousedown', handlePointer);
    window.addEventListener('scroll', onClose, true);
    window.addEventListener('resize', onClose);

    return () => {
      document.removeEventListener('mousedown', handlePointer);
      window.removeEventListener('scroll', onClose, true);
      window.removeEventListener('resize', onClose);
    };
  }, [open, onClose, refs]);
}

/**
 * Всплывающая панель, отрисованная порталом в document.body,
 * чтобы не обрезаться скроллящимися/overflow-контейнерами.
 */
function PopoverMenu({ open, anchorRect, popoverRef, children }) {
  if (!open || !anchorRect) return null;

  const style = {
    position: 'fixed',
    bottom: Math.max(12, window.innerHeight - anchorRect.top + 8),
    right: Math.max(12, window.innerWidth - anchorRect.right),
  };

  return createPortal(
    <div className="popover-menu" style={style} ref={popoverRef}>
      {children}
    </div>,
    document.body,
  );
}

function ScreenMetrics({ stats, label = 'LIVE', compact = false }) {
  return (
    <div
      className={`screen-metrics ${stats.ready ? 'is-ready' : ''} ${compact ? 'is-compact' : ''}`}
      aria-label="Статистика демонстрации экрана"
    >
      <div className="screen-metrics-live">
        <span className="screen-metrics-dot" aria-hidden="true" />
        {label}
      </div>

      <div className="screen-metric">
        <span>FPS</span>
        <strong>{stats.ready ? Math.round(stats.fps) : '—'}</strong>
      </div>

      <div className="screen-metric">
        <span>BITRATE</span>
        <strong>{stats.ready ? `${stats.bitrateMbps.toFixed(2)} Mbps` : '—'}</strong>
      </div>

      <div className="screen-metric">
        <span>RTT</span>
        <strong>
          {stats.ready && stats.rttMs > 0 ? `${Math.round(stats.rttMs)} ms` : '—'}
        </strong>
      </div>

      <div className="screen-metric screen-metric-secondary">
        <span>LOSS</span>
        <strong>{stats.ready ? `${stats.packetLoss.toFixed(1)}%` : '—'}</strong>
      </div>

      <div className="screen-metric screen-metric-secondary">
        <span>JITTER</span>
        <strong>
          {stats.ready && stats.jitterMs > 0 ? `${Math.round(stats.jitterMs)} ms` : '—'}
        </strong>
      </div>

      {stats.width > 0 && stats.height > 0 && (
        <div className="screen-metric screen-metric-secondary">
          <span>RES</span>
          <strong>{stats.width}×{stats.height}</strong>
        </div>
      )}
    </div>
  );
}

function VolumeSlider({ value, onChange, label }) {
  const fillPercent = Math.min(100, (value / VOLUME_MAX) * 100);

  return (
    <div className="volume-slider-row">
      <VolumeIcon volume={value} />

      <input
        type="range"
        min="0"
        max={VOLUME_MAX}
        step="1"
        value={value}
        onChange={(event) => onChange(Number(event.target.value))}
        className="volume-slider"
        style={{ '--volume-fill': `${fillPercent}%` }}
        aria-label={label}
      />

      <span className="volume-value">{value}%</span>
    </div>
  );
}

function ParticipantTile({
  participant,
  isSelf,
  isBeingViewed,
  onSelectDemo,
  screenAudioTrack,
  screenVolume,
  onScreenVolumeChange,
}) {
  const [volume, setVolume] = useState(100);
  const [mutedForMe, setMutedForMe] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [anchorRect, setAnchorRect] = useState(null);

  const buttonRef = useRef(null);
  const popoverRef = useRef(null);
  const menuRefs = useMemo(() => [buttonRef, popoverRef], []);

  const micEnabled = participant.isMicrophoneEnabled;
  const isSharingScreen = participant.isScreenShareEnabled;

  // Громкость применяется только локально для текущего пользователя
  // и не влияет на то, что слышат остальные участники комнаты.
  useEffect(() => {
    if (isSelf) return;
    if (typeof participant.setVolume !== 'function') return;
    try {
      participant.setVolume(mutedForMe ? 0 : volume / 100);
    } catch (error) {
      console.error('Не удалось изменить громкость участника:', error);
    }
  }, [participant, isSelf, volume, mutedForMe]);

  // Звук демонстрации экрана этого участника продолжает играть в фоне, даже
  // если вы сейчас смотрите чью-то ещё демку — поэтому громкость применяется
  // независимо от того, открыта ли она у вас на экране прямо сейчас.
  useEffect(() => {
    if (!screenAudioTrack || typeof screenAudioTrack.setVolume !== 'function') return;
    try {
      screenAudioTrack.setVolume(screenVolume / 100);
    } catch (error) {
      console.error('Не удалось изменить громкость демонстрации:', error);
    }
  }, [screenAudioTrack, screenVolume]);

  const closeMenu = useCallback(() => setMenuOpen(false), []);
  useOutsideDismiss(menuOpen, closeMenu, menuRefs);

  const toggleMenu = (event) => {
    event.stopPropagation();
    if (!menuOpen) {
      setAnchorRect(buttonRef.current.getBoundingClientRect());
    }
    setMenuOpen((prev) => !prev);
  };

  const displayName = participant.name || participant.identity;

  const statusText = isSelf
    ? 'Это вы'
    : mutedForMe
      ? 'Заглушён для вас'
      : participant.isSpeaking
        ? 'Говорит'
        : 'Слушает';

  return (
    <article
      className={`participant-card ${
        participant.isSpeaking ? 'speaking' : ''
      }`}
    >
      <Avatar name={displayName} size="lg" />

      <div>
        <strong>{displayName}</strong>
        <span>{statusText}</span>

        {isSharingScreen && (
          <button
            type="button"
            className={`share-indicator ${isBeingViewed ? 'is-active' : ''}`}
            onClick={() => onSelectDemo(participant.identity)}
          >
            <MonitorPlay size={12} />
            {isBeingViewed ? 'Демка открыта — свернуть' : 'Демонстрирует экран — смотреть'}
          </button>
        )}
      </div>

      <span
        className={`participant-status ${
          micEnabled ? 'mic-on' : 'mic-off'
        }`}
        title={micEnabled ? 'Микрофон включён' : 'Микрофон выключен'}
      >
        {micEnabled ? <Mic size={15} /> : <MicOff size={15} />}
      </span>

      {!isSelf && (
        <>
          <button
            ref={buttonRef}
            type="button"
            className={`participant-menu-trigger ${
              menuOpen || mutedForMe ? 'active' : ''
            }`}
            onClick={toggleMenu}
            aria-haspopup="true"
            aria-expanded={menuOpen}
            aria-label={`Настройки участника ${displayName}`}
            title="Настройки участника"
          >
            <MoreVertical size={16} />
          </button>

          <PopoverMenu
            open={menuOpen}
            anchorRect={anchorRect}
            popoverRef={popoverRef}
          >
            <div className="participant-popover-header">{displayName}</div>

            <button
              type="button"
              className={`popover-toggle-row ${mutedForMe ? 'is-on' : ''}`}
              onClick={() => setMutedForMe((prev) => !prev)}
            >
              <span>Выключить микрофон для себя</span>
              <span className="toggle-switch" aria-hidden="true">
                <span className="toggle-knob" />
              </span>
            </button>

            <div className="popover-divider" />

            <div
              className={`popover-volume ${mutedForMe ? 'is-disabled' : ''}`}
            >
              <span className="popover-label">Громкость участника</span>
              <VolumeSlider
                value={volume}
                onChange={setVolume}
                label={`Громкость участника ${displayName}`}
              />
            </div>

            {screenAudioTrack && (
              <>
                <div className="popover-divider" />

                <div className="popover-volume">
                  <span className="popover-label">Громкость демонстрации</span>
                  <VolumeSlider
                    value={screenVolume}
                    onChange={onScreenVolumeChange}
                    label={`Громкость демонстрации ${displayName}`}
                  />
                </div>
              </>
            )}
          </PopoverMenu>
        </>
      )}
    </article>
  );
}

function RoomInterior({ room, user, onLeave }) {
  const participants = useParticipants();
  const livekitRoom = useRoomContext();
  const { isMicrophoneEnabled, isScreenShareEnabled, localParticipant } = useLocalParticipant();

  const screenContainerRef = useRef(null);
  const [isFullscreen, setIsFullscreen] = useState(false);

  // Какую демонстрацию сейчас смотрит локальный пользователь. null — никакую
  // (свёрнуто в обычный вид комнаты), даже если кто-то продолжает делиться
  // экраном в фоне.
  const [viewedTrackKey, setViewedTrackKey] = useState(null);
  // true, если пользователь сам закрыл просмотр — тогда не переоткрываем
  // его автоматически при следующем обновлении списка демок.
  const manuallyClosedRef = useRef(false);

  // Громкость демонстрации хранится отдельно на каждого участника (по identity),
  // поэтому переключение между демками её не сбрасывает.
  const [screenVolumes, setScreenVolumes] = useState({});

  // Захват аудио при демонстрации всего экрана: хотим системный звук (другие
  // приложения, уведомления и т.п.), но не звук нашей же вкладки — иначе голоса
  // участников, которые уже играют у демонстрирующего, попадут обратно в его
  // демку и создадут эхо на всю комнату. restrictOwnAudio делает это точечно,
  // не отключая системный звук целиком. Поддерживается не всеми браузерами,
  // поэтому включаем его только если браузер о нём знает.
  const screenAudioOptions = useMemo(() => {
    const supportsRestrictOwnAudio =
      typeof navigator !== 'undefined' &&
      navigator.mediaDevices?.getSupportedConstraints?.().restrictOwnAudio;

    return supportsRestrictOwnAudio ? { restrictOwnAudio: true } : true;
  }, []);

  const screenTracks = useTracks([
    {
      source: Track.Source.ScreenShare,
      withPlaceholder: false,
    },
  ]);

  const screenAudioTracks = useTracks([
    {
      source: Track.Source.ScreenShareAudio,
      withPlaceholder: false,
    },
  ]);

  useEffect(() => {
    if (screenTracks.length === 0) {
      setViewedTrackKey(null);
      manuallyClosedRef.current = false;
      return;
    }

    const stillExists = screenTracks.some(
      (trackRef) => trackKey(trackRef) === viewedTrackKey,
    );

    if (stillExists || manuallyClosedRef.current) return;

    const latest = screenTracks[screenTracks.length - 1];
    setViewedTrackKey(trackKey(latest));
  }, [screenTracks, viewedTrackKey]);

  const activeTrack = screenTracks.find(
    (trackRef) => trackKey(trackRef) === viewedTrackKey,
  );

  const activeScreenTrack = activeTrack?.publication?.track;
  const screenStats = useScreenShareStats(activeScreenTrack);

  const localScreenTrack = isScreenShareEnabled
    ? localParticipant?.getTrackPublication(Track.Source.ScreenShare)?.track
    : null;
  const localScreenStats = useScreenShareStats(localScreenTrack);

  const isViewingScreenShare = Boolean(activeTrack);
  const isViewingOwnScreen =
    isViewingScreenShare && activeTrack?.participant.identity === localParticipant?.identity;
  const otherActiveShares = Math.max(0, screenTracks.length - (isViewingScreenShare ? 1 : 0));

  const selectParticipantDemo = (identity) => {
    const trackRef = screenTracks.find((t) => t.participant.identity === identity);
    if (!trackRef) return;

    const key = trackKey(trackRef);

    if (viewedTrackKey === key) {
      manuallyClosedRef.current = true;
      setViewedTrackKey(null);
    } else {
      manuallyClosedRef.current = false;
      setViewedTrackKey(key);
    }
  };

  const closeDemoView = async () => {
    if (document.fullscreenElement === screenContainerRef.current) {
      try {
        await document.exitFullscreen();
      } catch (error) {
        console.error('Не удалось выйти из полноэкранного режима:', error);
      }
    }

    manuallyClosedRef.current = true;
    setViewedTrackKey(null);
  };

  const setScreenVolumeFor = (identity, value) => {
    setScreenVolumes((prev) => ({ ...prev, [identity]: value }));
  };

  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(document.fullscreenElement === screenContainerRef.current);
    };

    document.addEventListener('fullscreenchange', handleFullscreenChange);

    return () => {
      document.removeEventListener('fullscreenchange', handleFullscreenChange);
    };
  }, []);

  // Звуковой сигнал при входе/выходе участника из комнаты.
  useEffect(() => {
    if (!livekitRoom) return undefined;

    const handleParticipantConnected = () => playPresenceChime('join');
    const handleParticipantDisconnected = () => playPresenceChime('leave');

    livekitRoom.on(RoomEvent.ParticipantConnected, handleParticipantConnected);
    livekitRoom.on(RoomEvent.ParticipantDisconnected, handleParticipantDisconnected);

    return () => {
      livekitRoom.off(RoomEvent.ParticipantConnected, handleParticipantConnected);
      livekitRoom.off(RoomEvent.ParticipantDisconnected, handleParticipantDisconnected);
    };
  }, [livekitRoom]);

  const toggleFullscreen = async () => {
    const container = screenContainerRef.current;

    if (!container) return;

    try {
      if (document.fullscreenElement === container) {
        await document.exitFullscreen();
        return;
      }

      await container.requestFullscreen();
    } catch (error) {
      console.error('Не удалось открыть полноэкранный режим:', error);
    }
  };

  return (
    <div className="room-space">
      <header className="room-header">
        <div>
          <span className="eyebrow">Голосовая комната</span>
          <h1>{room.name}</h1>
        </div>

        <button
          className="secondary-button danger"
          onClick={onLeave}
          type="button"
        >
          <LogOut size={17} />
          Выйти
        </button>
      </header>

      <main
        className={`stage ${
          isViewingScreenShare ? 'stage-sharing' : ''
        }`}
      >
        {!isViewingScreenShare && (
          <div className="stage-copy">
            <span className="stage-badge">
              <Users size={15} />
              {participants.length} онлайн
            </span>

            <h2>Вы в эфире</h2>

            <p>
              {screenTracks.length > 0
                ? 'Кто-то демонстрирует экран — нажмите на аватар участника со значком экрана, чтобы посмотреть.'
                : 'Говорите свободно. Здесь нет лишних элементов — только люди и голос.'}
            </p>
          </div>
        )}

        {isViewingScreenShare && (
          <section
            ref={screenContainerRef}
            className={`screen-share-stage ${
              isFullscreen ? 'is-fullscreen' : ''
            }`}
          >
            <div className="screen-share-toolbar">
              <div className="screen-share-label">
                <MonitorUp size={15} />
                <span>
                  Демонстрация экрана
                  {' — '}
                  {activeTrack?.participant.name ||
                    activeTrack?.participant.identity}
                </span>
                {otherActiveShares > 0 && (
                  <span
                    className="screen-share-more-hint"
                    title="Нажмите «Демонстрирует экран» на карточке другого участника, чтобы переключиться"
                  >
                    +{otherActiveShares}
                  </span>
                )}
              </div>

              <div className="screen-share-toolbar-actions">
                <button
                  className="screen-fullscreen-button"
                  type="button"
                  onClick={toggleFullscreen}
                  title={
                    isFullscreen
                      ? 'Выйти из полноэкранного режима'
                      : 'Открыть на весь экран'
                  }
                  aria-label={
                    isFullscreen
                      ? 'Выйти из полноэкранного режима'
                      : 'Открыть демонстрацию на весь экран'
                  }
                >
                  {isFullscreen ? (
                    <Minimize2 size={19} />
                  ) : (
                    <Maximize2 size={19} />
                  )}
                </button>

                <button
                  className="screen-close-button"
                  type="button"
                  onClick={closeDemoView}
                  title="Закрыть демонстрацию"
                  aria-label="Закрыть демонстрацию экрана"
                >
                  <X size={19} />
                </button>
              </div>
            </div>

            <div className="screen-share-content">
              {activeTrack && (
                <>
                  <VideoTrack
                    key={viewedTrackKey}
                    trackRef={activeTrack}
                    className="screen-share-video"
                  />

                  <ScreenMetrics
                    stats={screenStats}
                    label={isViewingOwnScreen ? 'ВАША ДЕМКА' : 'LIVE'}
                  />
                </>
              )}
            </div>
          </section>
        )}

        {isScreenShareEnabled && localScreenTrack && !isViewingOwnScreen && (
          <div className="self-screen-metrics-panel">
            <ScreenMetrics
              stats={localScreenStats}
              label="ВАША ДЕМКА"
              compact
            />
          </div>
        )}

        <div className="participant-grid">
          {participants.map((participant) => {
            const screenAudioTrackRef = screenAudioTracks.find(
              (trackRef) => trackRef.participant.identity === participant.identity,
            );

            return (
              <ParticipantTile
                key={participant.identity}
                participant={participant}
                isSelf={participant.identity === user.id}
                isBeingViewed={
                  isViewingScreenShare &&
                  activeTrack?.participant.identity === participant.identity
                }
                onSelectDemo={selectParticipantDemo}
                screenAudioTrack={screenAudioTrackRef?.publication?.track}
                screenVolume={screenVolumes[participant.identity] ?? 100}
                onScreenVolumeChange={(value) =>
                  setScreenVolumeFor(participant.identity, value)
                }
              />
            );
          })}
        </div>
      </main>

      <footer className="room-controls">
        <div className="self-chip">
          <Avatar
            name={user.username}
            src={user.avatar}
            size="sm"
          />

          <span>
            <strong>{user.username}</strong>
            <small>Подключён</small>
          </span>
        </div>

        <div className="room-control-buttons">
          <TrackToggle
            className="mic-toggle"
            source={Track.Source.Microphone}
            showIcon={false}
            captureOptions={{
              echoCancellation: true,
              noiseSuppression: true,
              autoGainControl: true,
            }}
          >
            {isMicrophoneEnabled ? <Mic size={20} /> : <MicOff size={20} />}
          </TrackToggle>

          <TrackToggle
            className="screen-toggle"
            source={Track.Source.ScreenShare}
            showIcon={false}
            captureOptions={{
              audio: screenAudioOptions,
              // Всегда запрашиваем 1080p30. При нехватке канала WebRTC
              // старается сохранить 30 FPS, снижая качество/битрейт картинки.
              resolution: {
                width: 1920,
                height: 1080,
                frameRate: 30,
              },
              contentHint: 'motion',
              // Не даём выбрать саму вкладку с этим приложением источником показа.
              selfBrowserSurface: 'exclude',
            }}
            publishOptions={{
              screenShareEncoding: {
                maxBitrate: 5_000_000,
                maxFramerate: 30,
                priority: 'high',
              },
              // При просадках канала WebRTC в первую очередь уменьшает качество,
              // а не резко превращает 30 FPS в слайд-шоу.
              degradationPreference: 'maintain-framerate',
              simulcast: true,
            }}
            onDeviceError={(error) => {
              console.error('Ошибка демонстрации экрана:', error);
            }}
          >
            {isScreenShareEnabled ? (
              <MonitorX size={20} />
            ) : (
              <MonitorUp size={20} />
            )}
          </TrackToggle>
        </div>

        <span className="control-hint">
          {isMicrophoneEnabled ? 'Микрофон включён' : 'Микрофон выключен'}
          {' · '}
          {isScreenShareEnabled ? 'Идёт показ экрана' : 'Показ экрана выключен'}
        </span>
      </footer>

      <RoomAudioRenderer />
    </div>
  );
}

export default function RoomExperience({
  room,
  user,
  connection,
  onLeave,
}) {
  return (
    <LiveKitRoom
      serverUrl={connection.url}
      token={connection.token}
      connect
      audio={false}
      video={false}
      options={{ webAudioMix: true }}
      className="livekit-root"
      onDisconnected={onLeave}
    >
      <RoomInterior
        room={room}
        user={user}
        onLeave={onLeave}
      />
    </LiveKitRoom>
  );
}
