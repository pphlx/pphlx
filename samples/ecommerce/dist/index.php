<?php // Establish database mock state (simulated PHP backend)
$productName = "MacBook Pro";
$productDesc = "Our most advanced series of chips ever built for a pro laptop. Each chip in the M5 family delivers phenomenal CPU performance, faster unified memory, and up to twice as fast SSD storage, allowing you to fly through AI tasks at mind-bending speeds.";
$priceUSD = 2599.00;
$availableStock = 3;
$initialStars = 5;
$badgeText = "Pro Choice";
$imageUrl = "assets/mac-macbook-pro.webp"; ?>

<!DOCTYPE html>
<html>
<head>
    <title>PPHLX Preview</title>
    <script src="https://unpkg.com/react@18/umd/react.development.js" crossorigin></script>
    <script src="https://unpkg.com/react-dom@18/umd/react-dom.development.js" crossorigin></script>
    <script src="https://unpkg.com/solid-js@1.8.0/dist/solid.js"></script>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="assets/css/app.css">
</head>
<body class="bg-[#080a0d] min-h-screen flex items-center justify-center p-6">
    
  <!-- Multi-Framework Hydrated Product Card (React + Solid + Svelte + PHP) -->
  <div class="max-w-sm mx-auto bg-[#0f1115]/95 border border-[#1f2430] rounded-3xl shadow-[0_20px_50px_rgba(0,0,0,0.5)] overflow-hidden relative group hover:border-[#4bf3c8]/50 transition-all duration-500">
    
    <!-- React Component: ProductImage (Handles product image and badge) -->
    <div id="pphlx-productimage-1784135546164204200" class="pphlx-island" data-component="ProductImage" data-framework="react" data-hydrate="load"></div>
<script>
  window.pphlxProps = window.pphlxProps || {};
  window.pphlxProps["pphlx-productimage-1784135546164204200"] = {"url": <?php echo json_encode($imageUrl); ?>,"badge": <?php echo json_encode($badgeText); ?>};
</script>


    <!-- PHP server-side data interpolation -->
    <div class="p-6">
      <div class="flex items-center justify-between mb-3.5">
        <span class="text-[10px] font-bold uppercase tracking-wider text-[#4bf3c8] bg-[#4bf3c8]/10 border border-[#4bf3c8]/20 px-2.5 py-1 rounded-full shadow-[0_0_12px_rgba(75,243,200,0.1)]">Developer Edition</span>
        <span class="text-xs font-medium text-gray-400 font-mono">Stock: <span class="text-[#4bf3c8]"><?php echo $availableStock; ?></span> left</span>
      </div>

      <h2 class="text-xl font-bold text-white mb-2 font-sans tracking-tight hover:text-[#4bf3c8] transition-colors"><?php echo $productName; ?></h2>
      <p class="text-xs text-gray-400 leading-relaxed mb-4 font-sans"><?php echo $productDesc; ?></p>
      
      <!-- Solid Component: StarRating (Interactive star ratings widget) -->
      <div id="pphlx-starrating-1784135546164204200" class="pphlx-island" data-component="StarRating" data-framework="solid" data-hydrate="load"></div>
<script>
  window.pphlxProps = window.pphlxProps || {};
  window.pphlxProps["pphlx-starrating-1784135546164204200"] = {"initialRating": <?php echo json_encode($initialStars); ?>};
</script>


      <div class="h-[1px] bg-gradient-to-r from-transparent via-gray-800 to-transparent my-4"></div>
      
      <div class="flex items-center justify-between mt-4">
        <div>
          <p class="text-[9px] text-gray-500 uppercase font-semibold tracking-wider">Special Offer Price</p>
          <p class="text-2xl font-extrabold text-white font-sans tracking-tight">$<?php echo number_format($priceUSD, 2); ?></p>
        </div>
        
        <!-- Svelte Component: BuyButton (Add-to-cart button tracking count) -->
        <div id="pphlx-buybutton-1784135546163685100" class="pphlx-island" data-component="BuyButton" data-framework="svelte" data-hydrate="load"></div>
<script>
  window.pphlxProps = window.pphlxProps || {};
  window.pphlxProps["pphlx-buybutton-1784135546163685100"] = {"stock": <?php echo json_encode($availableStock); ?>};
</script>

      </div>
    </div>
  </div>

    <script src="assets/js/app.js"></script>
</body>
</html>

