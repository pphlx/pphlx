import { createSignal, For } from 'solid-js';
export default function StarRating({ initialRating }) {
    const [rating, setRating] = createSignal(initialRating);
    return (
        <div class="flex items-center gap-1.5 my-3 select-none">
            <For each={[1, 2, 3, 4, 5]}>
                {(star) => (
                    <button onClick={() => setRating(star)} class="focus:outline-none transition-transform hover:scale-125 active:scale-95">
                        <svg class="w-4 h-4 transition-all" fill={star <= rating() ? "#4bf3c8" : "none"} stroke="#4bf3c8" stroke-width="1.8" style={star <= rating() ? "filter: drop-shadow(0 0 4px rgba(75, 243, 200, 0.45))" : ""} viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.907c.969 0 1.371 1.24.588 1.81l-3.97 2.883a1 1 0 00-.364 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.971-2.883a1 1 0 00-1.18 0l-3.97 2.883c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.364-1.118l-3.97-2.883c-.783-.57-.38-1.81.588-1.81h4.906a1 1 0 00.95-.69l1.519-4.674z" />
                        </svg>
                    </button>
                )}
            </For>
            <span class="text-[#4bf3c8]/80 text-[10px] ml-1.5 px-1.5 py-0.5 rounded bg-[#4bf3c8]/10 border border-[#4bf3c8]/20 font-mono">({rating()}/5)</span>
        </div>
    );
}
