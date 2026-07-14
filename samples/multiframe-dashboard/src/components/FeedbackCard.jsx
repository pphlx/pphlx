import React, { useState } from 'react';

export default function FeedbackCard(props) {
  const [rating, setRating] = useState(null);
  const [comment, setComment] = useState("");
  const [submitted, setSubmitted] = useState(false);

  const ratings = [
    { value: 1, label: "😠 Bad" },
    { value: 2, label: "😐 Okay" },
    { value: 3, label: "😊 Good" },
    { value: 4, label: "🚀 Excellent" }
  ];

  return (
    <div className="p-6 border border-[#54b9ff] rounded-lg bg-gray-900 shadow-md">
      <h3 className="text-xl font-bold text-[#54b9ff] mb-3">3. React: {props.title || "Feedback Module"}</h3>
      {!submitted ? (
        <div>
          <p className="text-sm text-gray-300 mb-3">Rate your experience with PPHLX:</p>
          <div className="flex gap-2.5 mb-4">
            {ratings.map(r => (
              <button
                key={r.value}
                onClick={() => setRating(r.value)}
                className={`px-3 py-1.5 rounded text-xs font-semibold cursor-pointer transition-all border-none ${
                  rating === r.value ? 'bg-[#54b9ff] text-black font-bold' : 'bg-gray-800 text-white hover:bg-gray-700'
                }`}
              >
                {r.label}
              </button>
            ))}
          </div>
          <textarea
            placeholder="Tell us what you think..."
            value={comment}
            onChange={e => setComment(e.target.value)}
            className="w-full bg-gray-800 border border-gray-700 text-white p-2.5 rounded text-xs min-h-[70px] mb-3 focus:outline-none focus:border-[#54b9ff]"
          />
          <button
            onClick={() => setSubmitted(true)}
            disabled={!rating}
            className={`px-4 py-2 rounded text-xs font-bold transition-all border-none ${
              rating ? 'bg-[#54b9ff] text-black cursor-pointer' : 'bg-gray-800 text-gray-500 cursor-not-allowed'
            }`}
          >
            Submit Feedback
          </button>
        </div>
      ) : (
        <div className="text-center py-2 text-sm text-gray-300">
          <h4 className="font-bold text-white mb-2">Thank you for your feedback!</h4>
          <p className="mb-1">Rating: <strong className="text-[#54b9ff]">{ratings.find(r => r.value === rating)?.label}</strong></p>
          {comment && <p className="italic my-2">"{comment}"</p>}
          <button onClick={() => setSubmitted(false)} className="mt-2 px-3 py-1 bg-gray-800 hover:bg-gray-700 text-[#54b9ff] border border-[#54b9ff]/30 rounded text-xs transition cursor-pointer">
            Rate Again
          </button>
        </div>
      )}
    </div>
  );
}
