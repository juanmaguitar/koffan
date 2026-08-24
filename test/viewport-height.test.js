const test = require('node:test');
const assert = require('node:assert/strict');

const {
    getVisibleViewportHeight,
    getKeyboardInset,
    syncViewportHeight
} = require('../static/viewport.js');

test('getVisibleViewportHeight prefers visualViewport height when available', () => {
    const win = {
        innerHeight: 780,
        visualViewport: {
            height: 512
        }
    };

    assert.equal(getVisibleViewportHeight(win), 512);
});

test('getVisibleViewportHeight falls back to innerHeight when visualViewport is unavailable', () => {
    const win = {
        innerHeight: 780
    };

    assert.equal(getVisibleViewportHeight(win), 780);
});

test('syncViewportHeight writes the viewport css variable in pixels', () => {
    const doc = {
        documentElement: {
            style: {
                values: {},
                setProperty(name, value) {
                    this.values[name] = value;
                }
            }
        }
    };

    const win = {
        innerHeight: 900,
        visualViewport: {
            height: 640
        }
    };

    const height = syncViewportHeight(win, doc);

    assert.equal(height, 640);
    assert.equal(
        doc.documentElement.style.values['--app-viewport-height'],
        '640px'
    );
});

test('getKeyboardInset reports the pixels the keyboard covers', () => {
    const win = {
        innerHeight: 900,
        visualViewport: {
            height: 560,
            offsetTop: 0
        }
    };

    assert.equal(getKeyboardInset(win), 340);
});

test('getKeyboardInset discounts the visual viewport offset when the page is scrolled', () => {
    const win = {
        innerHeight: 900,
        visualViewport: {
            height: 560,
            offsetTop: 40
        }
    };

    assert.equal(getKeyboardInset(win), 300);
});

test('getKeyboardInset is zero with no keyboard and never goes negative', () => {
    assert.equal(getKeyboardInset({ innerHeight: 900, visualViewport: { height: 900 } }), 0);
    // Pinch-zooming can make the visual viewport taller than the layout one.
    assert.equal(getKeyboardInset({ innerHeight: 900, visualViewport: { height: 1200 } }), 0);
});

test('getKeyboardInset falls back to zero without visualViewport support', () => {
    assert.equal(getKeyboardInset({ innerHeight: 900 }), 0);
    assert.equal(getKeyboardInset(undefined), 0);
});

test('syncViewportHeight publishes the keyboard inset and open state', () => {
    const doc = {
        documentElement: {
            attrs: {},
            style: {
                values: {},
                setProperty(name, value) {
                    this.values[name] = value;
                }
            },
            setAttribute(name, value) {
                this.attrs[name] = value;
            }
        }
    };

    syncViewportHeight({ innerHeight: 900, visualViewport: { height: 560 } }, doc);

    assert.equal(doc.documentElement.style.values['--app-keyboard-inset'], '340px');
    assert.equal(doc.documentElement.attrs['data-keyboard-open'], 'true');

    syncViewportHeight({ innerHeight: 900, visualViewport: { height: 900 } }, doc);

    assert.equal(doc.documentElement.style.values['--app-keyboard-inset'], '0px');
    assert.equal(doc.documentElement.attrs['data-keyboard-open'], 'false');
});
