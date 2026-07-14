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
    <div class="p-6 border border-[#4f80c2] rounded-lg bg-gray-900 shadow-md">
      <h3 class="text-xl font-bold text-[#4f80c2] mb-3">6. SolidJS: {props.label || "Stopwatch"}</h3>
      <div class="text-3xl font-mono font-bold text-white mb-4">
        {(time() / 10).toFixed(1)}s
      </div>
      <div class="flex gap-2.5">
        <button onClick={startStop} class="px-4 py-2 bg-[#4f80c2] text-white rounded text-xs font-bold transition-all cursor-pointer border-none">
          {running() ? "Pause" : "Start"}
        </button>
        <button onClick={reset} class="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-[#4f80c2] border border-[#4f80c2]/30 rounded text-xs transition cursor-pointer">
          Reset
        </button>
      </div>
    </div>
  );
}
