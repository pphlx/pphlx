<script setup>
import { ref, onMounted, onUnmounted } from 'vue';

const container = ref(null);
let animationFrameId = null;
let geometry, material, renderer;

onMounted(() => {
  if (!window.THREE) return;
  const THREE = window.THREE;

  const scene = new THREE.Scene();
  const camera = new THREE.PerspectiveCamera(45, 1, 0.1, 100);
  camera.position.z = 6;

  renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
  renderer.setSize(180, 180);
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));

  if (container.value) {
    container.value.appendChild(renderer.domElement);
  }

  // Sphere
  geometry = new THREE.SphereGeometry(1.4, 32, 16);
  material = new THREE.MeshNormalMaterial({ wireframe: true });
  const sphere = new THREE.Mesh(geometry, material);
  scene.add(sphere);

  const animate = () => {
    sphere.rotation.y += 0.01;
    sphere.rotation.x += 0.005;
    renderer.render(scene, camera);
    animationFrameId = requestAnimationFrame(animate);
  };
  animate();
});

onUnmounted(() => {
  if (animationFrameId) {
    cancelAnimationFrame(animationFrameId);
  }
  if (renderer && renderer.domElement && renderer.domElement.parentNode) {
    renderer.domElement.parentNode.removeChild(renderer.domElement);
  }
  if (geometry) geometry.dispose();
  if (material) material.dispose();
});
</script>

<template>
  <div class="p-6 border border-[#42b983] rounded-lg bg-gray-900 shadow-md flex flex-col items-center">
    <h3 class="text-xl font-bold text-[#42b983] mb-3 w-full text-left select-none">13. Vue 3: TresJS Logo</h3>
    <div ref="container" class="w-[180px] h-[180px] flex items-center justify-center"></div>
    <p class="text-xs text-gray-400 mt-2">Vue-Three.js WebGL rotating wireframe Sphere</p>
  </div>
</template>
