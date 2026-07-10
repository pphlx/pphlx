import React, { useState } from 'react';

export default function MyCounter({ title, initialCount }) {
  const [count, setCount] = useState(parseInt(initialCount || 0, 10));

  return (
    <div style={{ padding: '20px', border: '2px solid #4F46E5', borderRadius: '8px', maxWidth: '300px', margin: '20px auto', textAlign: 'center', fontFamily: 'sans-serif' }}>
      <h3 style={{ margin: '0 0 10px 0', color: '#374151' }}>{title}</h3>
      <div style={{ fontSize: '36px', fontWeight: 'bold', margin: '15px 0', color: '#111827' }}>{count}</div>
      <button 
        onClick={() => setCount(count + 1)}
        style={{ padding: '10px 20px', fontSize: '16px', backgroundColor: '#4F46E5', color: '#fff', border: 'none', borderRadius: '4px', cursor: 'pointer', transition: 'background-color 0.2s' }}
        onMouseOver={(e) => e.target.style.backgroundColor = '#4338CA'}
        onMouseOut={(e) => e.target.style.backgroundColor = '#4F46E5'}
      >
        Increment
      </button>
    </div>
  );
}
