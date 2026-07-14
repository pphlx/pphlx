<script>
  import { onMount } from 'svelte';

  let container;

  onMount(() => {
    if (!window.THREE) return;
    const THREE = window.THREE;

    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(45, 1, 0.1, 100);
    camera.position.z = 5;

    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setSize(180, 180);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    
    container.appendChild(renderer.domElement);

    // Box
    const geometry = new THREE.BoxGeometry(1.6, 1.6, 1.6);
    const material = new THREE.MeshNormalMaterial({ wireframe: true });
    const cube = new THREE.Mesh(geometry, material);
    scene.add(cube);

    let frame;
    function animate() {
      cube.rotation.x += 0.01;
      cube.rotation.y += 0.01;
      renderer.render(scene, camera);
      frame = requestAnimationFrame(animate);
    }
    animate();

    return () => {
      cancelAnimationFrame(frame);
      if (renderer.domElement && renderer.domElement.parentNode) {
        renderer.domElement.parentNode.removeChild(renderer.domElement);
      }
      geometry.dispose();
      material.dispose();
    };
  });
</script>

<div class="p-6 border border-[#ff3e00] rounded-lg bg-gray-900 shadow-md flex flex-col items-center">
  <h3 class="text-xl font-bold text-[#ff3e00] mb-3 w-full text-left select-none">12. Svelte 4: Threlte Logo</h3>
  <div bind:this={container} class="w-[180px] h-[180px] flex items-center justify-center"></div>
  <p class="text-xs text-gray-400 mt-2">Wireframe 3D Cube built with Svelte-Three.js</p>
</div>
