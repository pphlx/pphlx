import React, { useState } from 'react';

export default function ComplexDashboard({ serverData, initialActiveTab }) {
  const [activeTab, setActiveTab] = useState(initialActiveTab || 'servers');
  const [searchQuery, setSearchQuery] = useState('');
  
  // Parse servers if passed as JSON string, otherwise use directly
  const items = typeof serverData === 'string' ? JSON.parse(serverData) : serverData || [];

  // Filter items based on query
  const filteredItems = items.filter(item => 
    item.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    item.ip.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div style={{ fontFamily: 'sans-serif', backgroundColor: '#f9fafb', border: '1px solid #e5e7eb', borderRadius: '8px', padding: '24px', maxWidth: '600px', margin: '20px auto', boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)' }}>
      <div style={{ display: 'flex', borderBottom: '1px solid #e5e7eb', marginBottom: '20px' }}>
        <button 
          onClick={() => setActiveTab('servers')}
          style={{ padding: '10px 20px', border: 'none', background: 'none', borderBottom: activeTab === 'servers' ? '2px solid #4F46E5' : 'none', fontWeight: activeTab === 'servers' ? 'bold' : 'normal', cursor: 'pointer', color: activeTab === 'servers' ? '#4F46E5' : '#4B5563' }}
        >
          Active Servers
        </button>
        <button 
          onClick={() => setActiveTab('usage')}
          style={{ padding: '10px 20px', border: 'none', background: 'none', borderBottom: activeTab === 'usage' ? '2px solid #4F46E5' : 'none', fontWeight: activeTab === 'usage' ? 'bold' : 'normal', cursor: 'pointer', color: activeTab === 'usage' ? '#4F46E5' : '#4B5563' }}
        >
          CPU/RAM Usage Chart
        </button>
      </div>

      {activeTab === 'servers' && (
        <div>
          <input 
            type="text" 
            placeholder="Search servers by name or IP..." 
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #d1d5db', marginBottom: '16px', boxSizing: 'border-box' }}
          />

          <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
            {filteredItems.map((server, idx) => (
              <li key={idx} style={{ display: 'flex', justifyContent: 'space-between', padding: '12px', borderBottom: '1px solid #f3f4f6', backgroundColor: '#fff', borderRadius: '6px', marginBottom: '8px', boxShadow: '0 1px 2px rgba(0,0,0,0.05)' }}>
                <div>
                  <div style={{ fontWeight: 'bold', color: '#111827' }}>{server.name}</div>
                  <div style={{ fontSize: '12px', color: '#6B7280' }}>{server.ip}</div>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <span style={{ display: 'inline-block', padding: '4px 8px', borderRadius: '9999px', fontSize: '12px', fontWeight: 'bold', backgroundColor: server.status === 'Online' ? '#DEF7EC' : '#FDE8E8', color: server.status === 'Online' ? '#03543F' : '#9B1C1C' }}>
                    {server.status}
                  </span>
                </div>
              </li>
            ))}
            {filteredItems.length === 0 && (
              <li style={{ textAlign: 'center', color: '#9CA3AF', padding: '20px' }}>No servers found matching query.</li>
            )}
          </ul>
        </div>
      )}

      {activeTab === 'usage' && (
        <div>
          <h4 style={{ margin: '0 0 16px 0', color: '#374151' }}>Resource Usage Analytics</h4>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {items.map((server, idx) => (
              <div key={idx} style={{ backgroundColor: '#fff', padding: '12px', borderRadius: '6px', border: '1px solid #f3f4f6', boxShadow: '0 1px 2px rgba(0,0,0,0.05)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '14px', marginBottom: '6px' }}>
                  <span style={{ fontWeight: 'bold', color: '#374151' }}>{server.name}</span>
                  <span style={{ color: '#4B5563', fontWeight: 'bold' }}>{server.cpu}% CPU</span>
                </div>
                <div style={{ width: '100%', height: '8px', backgroundColor: '#e5e7eb', borderRadius: '9999px', overflow: 'hidden' }}>
                  <div style={{ width: `${server.cpu}%`, height: '100%', backgroundColor: server.cpu > 80 ? '#EF4444' : '#10B981', borderRadius: '9999px' }}></div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
