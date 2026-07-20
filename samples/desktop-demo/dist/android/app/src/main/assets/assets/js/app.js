
// Require shim for browser CDN compatibility
window.require = window.require || function(mod) {
  if (mod === 'react') return window.React;
  if (mod === 'react-dom') return window.ReactDOM;
  return undefined;
};

// Process environment shim for Vue/Svelte compatibility
window.process = window.process || { env: { NODE_ENV: 'production' } };

// PPHLX Desktop App API Bridge
window.pphlx = window.pphlx || {};
window.pphlx.desktop = {
  openFileDialog: async function(options) {
    if (window.pphlxDesktopOpenFile) {
      return await window.pphlxDesktopOpenFile();
    }
    console.warn("pphlx.desktop.openFileDialog is only available in native desktop target.");
    return null;
  },
  saveFileDialog: async function(options) {
    if (window.pphlxDesktopSaveFile) {
      return await window.pphlxDesktopSaveFile();
    }
    console.warn("pphlx.desktop.saveFileDialog is only available in native desktop target.");
    return null;
  },
  showNotification: function(title, message) {
    if (window.pphlxDesktopShowNotification) {
      window.pphlxDesktopShowNotification(title, message);
    } else {
      alert(title + ": " + message);
    }
  },
  window: {
    minimize: function() {
      if (window.pphlxDesktopMinimize) {
        window.pphlxDesktopMinimize();
      } else {
        console.warn("window.minimize is only available in native desktop target.");
      }
    },
    maximize: function() {
      if (window.pphlxDesktopMaximize) {
        window.pphlxDesktopMaximize();
      } else {
        console.warn("window.maximize is only available in native desktop target.");
      }
    },
    close: function() {
      if (window.pphlxDesktopClose) {
        window.pphlxDesktopClose();
      } else {
        window.close();
      }
    }
  }
};

// PPHLX Islands Runtime
document.addEventListener("DOMContentLoaded", () => {
  document.querySelectorAll(".pphlx-island").forEach(island => {
    const compName = island.getAttribute("data-component");
    const framework = island.getAttribute("data-framework") || "";
    const islandId = island.id;
    const props = window.pphlxProps ? window.pphlxProps[islandId] : {};
    
    if (window[compName]) {
      const ComponentModule = window[compName];
      const Component = ComponentModule.default || ComponentModule;
      
      if (framework === "react" && window.ReactDOM && window.ReactDOM.createRoot) {
        // React 18+ Mount
        const root = window.ReactDOM.createRoot(island);
        root.render(window.React.createElement(Component, props));
      } else if (framework === "vue" && window.Vue && window.Vue.createApp) {
        // Vue 3 Mount
        window.Vue.createApp(Component, props).mount(island);
      } else if (framework === "svelte") {
        // Support Svelte 4 classes and Svelte 5 functions
        if (Component.prototype && Component.prototype.$destroy) {
          new Component({ target: island, props: props });
        } else if (window.Svelte && window.Svelte.mount) {
          window.Svelte.mount(Component, { target: island, props: props });
        } else if (typeof Component === "function") {
          Component(island, props);
        }
      } else if (framework === "solid" && window.SolidJS && window.SolidJS.render) {
        // SolidJS Mount
        window.SolidJS.render(() => Component(props), island);
      } else if (framework === "preact" && window.preact && window.preact.render) {
        // Preact Mount
        window.preact.render(window.preact.h(Component, props), island);
      } else if (Component.render) {
        Component.render(island, props);
      } else {
        // Backwards compatibility fallback if data-framework not set
        if (window.ReactDOM && window.ReactDOM.createRoot) {
          const root = window.ReactDOM.createRoot(island);
          root.render(window.React.createElement(Component, props));
        } else if (window.Vue && window.Vue.createApp) {
          window.Vue.createApp(Component, props).mount(island);
        } else if (typeof Component === "function") {
          if (Component.prototype && Component.prototype.$destroy) {
            new Component({ target: island, props: props });
          } else if (window.Svelte && window.Svelte.mount) {
            window.Svelte.mount(Component, { target: island, props: props });
          } else {
            Component(island, props);
          }
        } else {
          console.warn("No runtime renderer found for component " + compName);
        }
      }
    } else {
      console.error("Component " + compName + " not found in window scope.");
    }
  });
});

