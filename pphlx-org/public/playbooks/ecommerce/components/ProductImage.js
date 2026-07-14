import React from 'react';
export default function ProductImage({ url, badge }) {
    return (
        <div className="relative overflow-hidden rounded-t-3xl bg-gradient-to-b from-[#1b202e] to-[#0d0f14] h-52 flex items-center justify-center border-b border-[#1f2430]">
            <img src={url} alt="Product" className="object-cover w-4/5 h-auto hover:scale-105 transition-transform duration-500" />
            {badge && (
                <span className="absolute top-4 left-4 bg-[#4bf3c8]/15 border border-[#4bf3c8]/30 text-[#4bf3c8] backdrop-blur-md text-[9px] font-bold px-2.5 py-1 rounded-lg shadow-[0_0_15px_rgba(75,243,200,0.15)] uppercase tracking-wider">
                     {badge}
                </span>
            )}
        </div>
    );
}
