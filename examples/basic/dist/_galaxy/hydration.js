const islands = new Map();
const strategies = {
  load: (fn) => fn(),
  idle: (fn) => {
    if ('requestIdleCallback' in window) {
      requestIdleCallback(fn);
    } else {
      setTimeout(fn, 1);
    }
  },
  visible: (fn, element) => {
    const observer = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting) {
        fn();
        observer.disconnect();
      }
    });
    observer.observe(element);
  },
  media: (fn, query) => {
    const mql = window.matchMedia(query);
    if (mql.matches) {
      fn();
    } else {
      mql.addEventListener('change', fn, { once: true });
    }
  }
};

export function hydrate(id, componentPath, props, strategy) {
  const element = document.querySelector(`[data-island-id="${id}"]`);
  if (!element) return;

  const hydrateComponent = async () => {
    try {
      const module = await import(componentPath);
      const Component = module.default;
      
      if (Component && Component.hydrate) {
        Component.hydrate(element, props);
      }
      
      islands.set(id, { element, props, component: Component });
    } catch (err) {
      console.error(`Failed to hydrate island ${id}:`, err);
    }
  };

  const strategyFn = strategies[strategy] || strategies.load;
  strategyFn(hydrateComponent, element);
}

export function getIsland(id) {
  return islands.get(id);
}
