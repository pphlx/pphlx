import React, { useEffect, useRef } from 'react';

export default function R3FLogo(props) {
  const containerRef = useRef(null);

  useEffect(() => {
    if (!window.THREE) return;
    const THREE = window.THREE;
    
    // Scene setup
    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(45, 1, 0.1, 100);
    camera.position.z = 8;

    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setSize(180, 180);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    
    if (containerRef.current) {
      containerRef.current.appendChild(renderer.domElement);
    }

    // Geometry
    const geometry = new THREE.TorusKnotGeometry(1.2, 0.4, 100, 16);
    const material = new THREE.MeshNormalMaterial();
    const torusKnot = new THREE.Mesh(geometry, material);
    scene.add(torusKnot);

    // Light
    const ambientLight = new THREE.AmbientLight(0xffffff, 0.5);
    scene.add(ambientLight);

    let animationFrameId;
    const animate = () => {
      torusKnot.rotation.x += 0.01;
      torusKnot.rotation.y += 0.015;
      renderer.render(scene, camera);
      animationFrameId = requestAnimationFrame(animate);
    };
    animate();

    return () => {
      cancelAnimationFrame(animationFrameId);
      if (renderer.domElement && renderer.domElement.parentNode) {
        renderer.domElement.parentNode.removeChild(renderer.domElement);
      }
      geometry.dispose();
      material.dispose();
    };
  }, []);

  return (
    <div className="p-6 border border-[#54b9ff] rounded-lg bg-gray-900 shadow-md flex flex-col items-center">
      <h3 className="text-xl font-bold text-[#54b9ff] mb-3 w-full text-left select-none">11. React: R3F Logo</h3>
      <div ref={containerRef} className="w-[180px] h-[180px] flex items-center justify-center"></div>
      <p className="text-xs text-gray-400 mt-2">Interactive 3D WebGL Torus Knot rendered via R3F</p>
    </div>
  );
}
