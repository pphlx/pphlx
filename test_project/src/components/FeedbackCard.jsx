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
    <div className="feedback-card" style={{ padding: '20px', border: '1px solid #61dafb', borderRadius: '8px', background: '#1a1a24', color: '#fff', margin: '15px 0' }}>
      <h3 style={{ color: '#61dafb', marginTop: 0 }}>React: {props.title || "Feedback Module"}</h3>
      {!submitted ? (
        <div>
          <p>Rate your experience with Piplex:</p>
          <div style={{ display: 'flex', gap: '10px', marginBottom: '15px' }}>
            {ratings.map(r => (
              <button
                key={r.value}
                onClick={() => setRating(r.value)}
                style={{
                  background: rating === r.value ? '#61dafb' : '#2d2d3d',
                  color: rating === r.value ? '#000' : '#fff',
                  border: 'none', padding: '8px 12px', borderRadius: '4px', cursor: 'pointer', fontWeight: 'bold'
                }}
              >
                {r.label}
              </button>
            ))}
          </div>
          <textarea
            placeholder="Tell us what you think..."
            value={comment}
            onChange={e => setComment(e.target.value)}
            style={{ width: '100%', boxSizing: 'border-box', background: '#2d2d3d', border: '1px solid #444', borderRadius: '4px', padding: '10px', color: '#fff', minHeight: '80px', marginBottom: '10px' }}
          />
          <button
            onClick={() => setSubmitted(true)}
            disabled={!rating}
            style={{ background: rating ? '#61dafb' : '#444', color: '#000', border: 'none', padding: '10px 20px', borderRadius: '4px', cursor: rating ? 'pointer' : 'not-allowed', fontWeight: 'bold' }}
          >
            Submit Feedback
          </button>
        </div>
      ) : (
        <div style={{ textAlign: 'center', padding: '10px 0' }}>
          <h4>Thank you for your feedback!</h4>
          <p>Rating: <strong>{ratings.find(r => r.value === rating)?.label}</strong></p>
          {comment && <p>"{comment}"</p>}
          <button onClick={() => setSubmitted(false)} style={{ background: '#2d2d3d', color: '#61dafb', border: '1px solid #61dafb', padding: '6px 12px', borderRadius: '4px', cursor: 'pointer' }}>
            Rate Again
          </button>
        </div>
      )}
    </div>
  );
}
