import { createSignal, onCleanup } from 'solid-js';

export default function SolidStopwatch(props) {
  const [time, setTime] = createSignal(0);
  const [running, setRunning] = createSignal(false);
  let timer;

  const startStop = () => {
    if (running()) {
      clearInterval(timer);
      setRunning(false);
    } else {
      setRunning(true);
      timer = setInterval(() => {
        setTime((t) => t + 1);
      }, 100);
    }
  };

  const reset = () => {
    clearInterval(timer);
    setRunning(false);
    setTime(0);
  };

  onCleanup(() => clearInterval(timer));

  return (
    <div class="solid-card" style="padding: 20px; border: 1px solid #4f80c2; border-radius: 8px; background: #1a1a24; color: #fff; margin: 15px 0;">
      <h3 style="color: #4f80c2; margin-top: 0;">SolidJS: {props.label || "Stopwatch"}</h3>
      <div style="font-size: 2em; font-family: monospace; font-weight: bold; margin-bottom: 15px;">
        {(time() / 10).toFixed(1)}s
      </div>
      <div style="display: flex; gap: 10px;">
        <button onClick={startStop} style="background: #4f80c2; color: #fff; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; font-weight: bold;">
          {running() ? "Pause" : "Start"}
        </button>
        <button onClick={reset} style="background: #2d2d3d; color: #fff; border: 1px solid #4f80c2; padding: 8px 16px; border-radius: 4px; cursor: pointer;">
          Reset
        </button>
      </div>
    </div>
  );
}
