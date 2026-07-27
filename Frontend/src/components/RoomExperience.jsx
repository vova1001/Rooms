import {
  LiveKitRoom,
  RoomAudioRenderer,
  TrackToggle,
  VideoTrack,
  useParticipants,
  useTracks,
} from '@livekit/components-react';
import { Track } from 'livekit-client';
import {
  LogOut,
  Maximize2,
  Mic,
  Minimize2,
  MonitorUp,
  Users,
} from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import Avatar from './Avatar';

function RoomInterior({ room, user, onLeave }) {
  const participants = useParticipants();
  const screenContainerRef = useRef(null);
  const [isFullscreen, setIsFullscreen] = useState(false);

  const screenTracks = useTracks([
    {
      source: Track.Source.ScreenShare,
      withPlaceholder: false,
    },
  ]);

  const isScreenSharing = screenTracks.length > 0;

  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(document.fullscreenElement === screenContainerRef.current);
    };

    document.addEventListener('fullscreenchange', handleFullscreenChange);

    return () => {
      document.removeEventListener('fullscreenchange', handleFullscreenChange);
    };
  }, []);

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
          isScreenSharing ? 'stage-sharing' : ''
        }`}
      >
        {!isScreenSharing && (
          <div className="stage-copy">
            <span className="stage-badge">
              <Users size={15} />
              {participants.length} онлайн
            </span>

            <h2>Вы в эфире</h2>

            <p>
              Говорите свободно. Здесь нет лишних элементов — только люди и голос.
            </p>
          </div>
        )}

        {isScreenSharing && (
          <section
            ref={screenContainerRef}
            className={`screen-share-stage ${
              isFullscreen ? 'is-fullscreen' : ''
            }`}
          >
            <div className="screen-share-toolbar">
              <div className="screen-share-label">
                <MonitorUp size={15} />
                <span>Демонстрация экрана</span>
              </div>

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
            </div>

            <div className="screen-share-content">
              {screenTracks.map((trackRef) => (
                <VideoTrack
                  key={
                    trackRef.publication?.trackSid ||
                    `${trackRef.participant.identity}-screen`
                  }
                  trackRef={trackRef}
                  className="screen-share-video"
                />
              ))}
            </div>
          </section>
        )}

        <div className="participant-grid">
          {participants.map((participant) => (
            <article
              className={`participant-card ${
                participant.isSpeaking ? 'speaking' : ''
              }`}
              key={participant.identity}
            >
              <Avatar
                name={participant.name || participant.identity}
                size="lg"
              />

              <div>
                <strong>
                  {participant.name || participant.identity}
                </strong>

                <span>
                  {participant.identity === user.id
                    ? 'Это вы'
                    : participant.isSpeaking
                      ? 'Говорит'
                      : 'Слушает'}
                </span>
              </div>

              <span className="participant-status">
                <Mic size={15} />
              </span>
            </article>
          ))}
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
          >
            <Mic size={20} />
          </TrackToggle>

          <TrackToggle
            className="screen-toggle"
            source={Track.Source.ScreenShare}
            onDeviceError={(error) => {
              console.error('Ошибка демонстрации экрана:', error);
            }}
          >
            <MonitorUp size={20} />
          </TrackToggle>
        </div>

        <span className="control-hint">
          Микрофон и демонстрация
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