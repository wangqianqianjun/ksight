/*
 _       __      _ __
| |     / /___ _(_) /____
| | /| / / __ `/ / / ___/
| |/ |/ / /_/ / / (__  )
|__/|__/\__,_/_/_/____/
The electron alternative for Go
(c) Lea Anthony 2019-present
*/

// Helper to check if Wails runtime is available
function isWailsAvailable() {
    return typeof window !== 'undefined' && window.runtime;
}

// Log once that dev mode is active
let devModeLogged = false;
function logDevMode() {
    if (!devModeLogged) {
        console.info('[Wails] Running in dev mode - runtime calls are no-ops');
        devModeLogged = true;
    }
}

export function LogPrint(message) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.LogPrint(message);
}

export function LogTrace(message) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.LogTrace(message);
}

export function LogDebug(message) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.LogDebug(message);
}

export function LogInfo(message) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.LogInfo(message);
}

export function LogWarning(message) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.LogWarning(message);
}

export function LogError(message) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.LogError(message);
}

export function LogFatal(message) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.LogFatal(message);
}

export function EventsOnMultiple(eventName, callback, maxCallbacks) {
    if (!isWailsAvailable()) { logDevMode(); return () => {}; }
    return window.runtime.EventsOnMultiple(eventName, callback, maxCallbacks);
}

export function EventsOn(eventName, callback) {
    return EventsOnMultiple(eventName, callback, -1);
}

export function EventsOff(eventName, ...additionalEventNames) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    return window.runtime.EventsOff(eventName, ...additionalEventNames);
}

export function EventsOnce(eventName, callback) {
    return EventsOnMultiple(eventName, callback, 1);
}

export function EventsEmit(eventName) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    const args = [eventName].slice.call(arguments);
    return window.runtime.EventsEmit.apply(null, args);
}

export function WindowReload() {
    if (!isWailsAvailable()) { logDevMode(); window.location.reload(); return; }
    window.runtime.WindowReload();
}

export function WindowReloadApp() {
    if (!isWailsAvailable()) { logDevMode(); window.location.reload(); return; }
    window.runtime.WindowReloadApp();
}

export function WindowSetAlwaysOnTop(b) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowSetAlwaysOnTop(b);
}

export function WindowSetSystemDefaultTheme() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowSetSystemDefaultTheme();
}

export function WindowSetLightTheme() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowSetLightTheme();
}

export function WindowSetDarkTheme() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowSetDarkTheme();
}

export function WindowCenter() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowCenter();
}

export function WindowSetTitle(title) {
    if (!isWailsAvailable()) { logDevMode(); document.title = title; return; }
    window.runtime.WindowSetTitle(title);
}

export function WindowFullscreen() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowFullscreen();
}

export function WindowUnfullscreen() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowUnfullscreen();
}

export function WindowIsFullscreen() {
    if (!isWailsAvailable()) { logDevMode(); return false; }
    return window.runtime.WindowIsFullscreen();
}

export function WindowGetSize() {
    if (!isWailsAvailable()) { logDevMode(); return { w: window.innerWidth, h: window.innerHeight }; }
    return window.runtime.WindowGetSize();
}

export function WindowSetSize(width, height) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowSetSize(width, height);
}

export function WindowSetMaxSize(width, height) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowSetMaxSize(width, height);
}

export function WindowSetMinSize(width, height) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowSetMinSize(width, height);
}

export function WindowSetPosition(x, y) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowSetPosition(x, y);
}

export function WindowGetPosition() {
    if (!isWailsAvailable()) { logDevMode(); return { x: 0, y: 0 }; }
    return window.runtime.WindowGetPosition();
}

export function WindowHide() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowHide();
}

export function WindowShow() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowShow();
}

export function WindowMaximise() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowMaximise();
}

export function WindowToggleMaximise() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowToggleMaximise();
}

export function WindowUnmaximise() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowUnmaximise();
}

export function WindowIsMaximised() {
    if (!isWailsAvailable()) { logDevMode(); return false; }
    return window.runtime.WindowIsMaximised();
}

export function WindowMinimise() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowMinimise();
}

export function WindowUnminimise() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowUnminimise();
}

export function WindowSetBackgroundColour(R, G, B, A) {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.WindowSetBackgroundColour(R, G, B, A);
}

export function ScreenGetAll() {
    if (!isWailsAvailable()) { logDevMode(); return []; }
    return window.runtime.ScreenGetAll();
}

export function WindowIsMinimised() {
    if (!isWailsAvailable()) { logDevMode(); return false; }
    return window.runtime.WindowIsMinimised();
}

export function WindowIsNormal() {
    if (!isWailsAvailable()) { logDevMode(); return true; }
    return window.runtime.WindowIsNormal();
}

export function BrowserOpenURL(url) {
    if (!isWailsAvailable()) { logDevMode(); window.open(url, '_blank'); return; }
    window.runtime.BrowserOpenURL(url);
}

export function Environment() {
    if (!isWailsAvailable()) { logDevMode(); return { buildType: 'dev', platform: 'browser', arch: 'unknown' }; }
    return window.runtime.Environment();
}

export function Quit() {
    if (!isWailsAvailable()) { logDevMode(); console.log('[Wails] Quit called in dev mode'); return; }
    window.runtime.Quit();
}

export function Hide() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.Hide();
}

export function Show() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    window.runtime.Show();
}

export function ClipboardGetText() {
    if (!isWailsAvailable()) { logDevMode(); return navigator.clipboard.readText(); }
    return window.runtime.ClipboardGetText();
}

export function ClipboardSetText(text) {
    if (!isWailsAvailable()) { logDevMode(); return navigator.clipboard.writeText(text); }
    return window.runtime.ClipboardSetText(text);
}

/**
 * Callback for OnFileDrop returns a slice of file path strings when a drop is finished.
 *
 * @export
 * @callback OnFileDropCallback
 * @param {number} x - x coordinate of the drop
 * @param {number} y - y coordinate of the drop
 * @param {string[]} paths - A list of file paths.
 */

/**
 * OnFileDrop listens to drag and drop events and calls the callback with the coordinates of the drop and an array of path strings.
 *
 * @export
 * @param {OnFileDropCallback} callback - Callback for OnFileDrop returns a slice of file path strings when a drop is finished.
 * @param {boolean} [useDropTarget=true] - Only call the callback when the drop finished on an element that has the drop target style. (--wails-drop-target)
 */
export function OnFileDrop(callback, useDropTarget) {
    if (!isWailsAvailable()) { logDevMode(); return () => {}; }
    return window.runtime.OnFileDrop(callback, useDropTarget);
}

/**
 * OnFileDropOff removes the drag and drop listeners and handlers.
 */
export function OnFileDropOff() {
    if (!isWailsAvailable()) { logDevMode(); return; }
    return window.runtime.OnFileDropOff();
}

export function CanResolveFilePaths() {
    if (!isWailsAvailable()) { logDevMode(); return false; }
    return window.runtime.CanResolveFilePaths();
}

export function ResolveFilePaths(files) {
    if (!isWailsAvailable()) { logDevMode(); return Promise.resolve([]); }
    return window.runtime.ResolveFilePaths(files);
}
