(function(root, factory) {
    const api = factory();

    if (typeof module === 'object' && module.exports) {
        module.exports = api;
    }

    root.ViewportHeight = api;
})(typeof globalThis !== 'undefined' ? globalThis : this, function() {
    function getVisibleViewportHeight(win) {
        const visualHeight = win?.visualViewport?.height;
        if (typeof visualHeight === 'number' && visualHeight > 0) {
            return visualHeight;
        }

        const innerHeight = win?.innerHeight;
        if (typeof innerHeight === 'number' && innerHeight > 0) {
            return innerHeight;
        }

        return 0;
    }

    // How many pixels at the bottom of the layout viewport are covered by the
    // on-screen keyboard. Position:fixed is relative to the layout viewport, so
    // a bar pinned to bottom:0 hides behind the keyboard without this offset.
    function getKeyboardInset(win) {
        const vv = win?.visualViewport;
        const layoutHeight = win?.innerHeight;

        if (!vv || typeof layoutHeight !== 'number' || typeof vv.height !== 'number') {
            return 0;
        }

        const inset = layoutHeight - vv.height - (vv.offsetTop || 0);
        return inset > 0 ? inset : 0;
    }

    function syncViewportHeight(win, doc) {
        const height = Math.round(getVisibleViewportHeight(win));
        if (!height || !doc?.documentElement?.style?.setProperty) {
            return height;
        }

        const keyboardInset = Math.round(getKeyboardInset(win));

        doc.documentElement.style.setProperty('--app-viewport-height', `${height}px`);
        doc.documentElement.style.setProperty('--app-keyboard-inset', `${keyboardInset}px`);
        doc.documentElement.setAttribute?.('data-keyboard-open', keyboardInset > 0 ? 'true' : 'false');
        return height;
    }

    function initViewportHeightSync(win, doc) {
        if (!win || !doc) {
            return function noop() {};
        }

        const sync = () => syncViewportHeight(win, doc);
        const removeListeners = [];

        const register = (target, eventName) => {
            if (!target?.addEventListener) return;
            target.addEventListener(eventName, sync, { passive: true });
            removeListeners.push(() => target.removeEventListener(eventName, sync));
        };

        sync();
        register(win, 'resize');
        register(win, 'orientationchange');
        register(win.visualViewport, 'resize');
        register(win.visualViewport, 'scroll');

        return function cleanup() {
            removeListeners.forEach(removeListener => removeListener());
        };
    }

    return {
        getVisibleViewportHeight,
        getKeyboardInset,
        syncViewportHeight,
        initViewportHeightSync
    };
});
