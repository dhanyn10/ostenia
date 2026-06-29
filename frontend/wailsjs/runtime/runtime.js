/*
const _appWindow = window; // NOSONAR
 _       __      _ __
| |     / /___ _(_) /____
| | /| / / __ `/ / / ___/
| |/ |/ / /_/ / / (__  )
|__/|__/\__,_/_/_/____/
The electron alternative for Go
(c) Lea Anthony 2019-present
*/

export function LogPrint(message) {
    _appWindow.runtime.LogPrint(message);
}

export function LogTrace(message) {
    _appWindow.runtime.LogTrace(message);
}

export function LogDebug(message) {
    _appWindow.runtime.LogDebug(message);
}

export function LogInfo(message) {
    _appWindow.runtime.LogInfo(message);
}

export function LogWarning(message) {
    _appWindow.runtime.LogWarning(message);
}

export function LogError(message) {
    _appWindow.runtime.LogError(message);
}

export function LogFatal(message) {
    _appWindow.runtime.LogFatal(message);
}

export function EventsOnMultiple(eventName, callback, maxCallbacks) {
    return _appWindow.runtime.EventsOnMultiple(eventName, callback, maxCallbacks);
}

export function EventsOn(eventName, callback) {
    return EventsOnMultiple(eventName, callback, -1);
}

export function EventsOff(eventName, ...additionalEventNames) {
    return _appWindow.runtime.EventsOff(eventName, ...additionalEventNames);
}

export function EventsOffAll() {
  return _appWindow.runtime.EventsOffAll();
}

export function EventsOnce(eventName, callback) {
    return EventsOnMultiple(eventName, callback, 1);
}

export function EventsEmit(eventName) {
    let args = [eventName].slice.call(arguments);
    return _appWindow.runtime.EventsEmit.apply(null, args);
}

export function WindowReload() {
    _appWindow.runtime.WindowReload();
}

export function WindowReloadApp() {
    _appWindow.runtime.WindowReloadApp();
}

export function WindowSetAlwaysOnTop(b) {
    _appWindow.runtime.WindowSetAlwaysOnTop(b);
}

export function WindowSetSystemDefaultTheme() {
    _appWindow.runtime.WindowSetSystemDefaultTheme();
}

export function WindowSetLightTheme() {
    _appWindow.runtime.WindowSetLightTheme();
}

export function WindowSetDarkTheme() {
    _appWindow.runtime.WindowSetDarkTheme();
}

export function WindowCenter() {
    _appWindow.runtime.WindowCenter();
}

export function WindowSetTitle(title) {
    _appWindow.runtime.WindowSetTitle(title);
}

export function WindowFullscreen() {
    _appWindow.runtime.WindowFullscreen();
}

export function WindowUnfullscreen() {
    _appWindow.runtime.WindowUnfullscreen();
}

export function WindowIsFullscreen() {
    return _appWindow.runtime.WindowIsFullscreen();
}

export function WindowGetSize() {
    return _appWindow.runtime.WindowGetSize();
}

export function WindowSetSize(width, height) {
    _appWindow.runtime.WindowSetSize(width, height);
}

export function WindowSetMaxSize(width, height) {
    _appWindow.runtime.WindowSetMaxSize(width, height);
}

export function WindowSetMinSize(width, height) {
    _appWindow.runtime.WindowSetMinSize(width, height);
}

export function WindowSetPosition(x, y) {
    _appWindow.runtime.WindowSetPosition(x, y);
}

export function WindowGetPosition() {
    return _appWindow.runtime.WindowGetPosition();
}

export function WindowHide() {
    _appWindow.runtime.WindowHide();
}

export function WindowShow() {
    _appWindow.runtime.WindowShow();
}

export function WindowMaximise() {
    _appWindow.runtime.WindowMaximise();
}

export function WindowToggleMaximise() {
    _appWindow.runtime.WindowToggleMaximise();
}

export function WindowUnmaximise() {
    _appWindow.runtime.WindowUnmaximise();
}

export function WindowIsMaximised() {
    return _appWindow.runtime.WindowIsMaximised();
}

export function WindowMinimise() {
    _appWindow.runtime.WindowMinimise();
}

export function WindowUnminimise() {
    _appWindow.runtime.WindowUnminimise();
}

export function WindowSetBackgroundColour(R, G, B, A) {
    _appWindow.runtime.WindowSetBackgroundColour(R, G, B, A);
}

export function ScreenGetAll() {
    return _appWindow.runtime.ScreenGetAll();
}

export function WindowIsMinimised() {
    return _appWindow.runtime.WindowIsMinimised();
}

export function WindowIsNormal() {
    return _appWindow.runtime.WindowIsNormal();
}

export function BrowserOpenURL(url) {
    _appWindow.runtime.BrowserOpenURL(url);
}

export function Environment() {
    return _appWindow.runtime.Environment();
}

export function Quit() {
    _appWindow.runtime.Quit();
}

export function Hide() {
    _appWindow.runtime.Hide();
}

export function Show() {
    _appWindow.runtime.Show();
}

export function ClipboardGetText() {
    return _appWindow.runtime.ClipboardGetText();
}

export function ClipboardSetText(text) {
    return _appWindow.runtime.ClipboardSetText(text);
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
    return _appWindow.runtime.OnFileDrop(callback, useDropTarget);
}

/**
 * OnFileDropOff removes the drag and drop listeners and handlers.
 */
export function OnFileDropOff() {
    return _appWindow.runtime.OnFileDropOff();
}

export function CanResolveFilePaths() {
    return _appWindow.runtime.CanResolveFilePaths();
}

export function ResolveFilePaths(files) {
    return _appWindow.runtime.ResolveFilePaths(files);
}

export function InitializeNotifications() {
    return _appWindow.runtime.InitializeNotifications();
}

export function CleanupNotifications() {
    return _appWindow.runtime.CleanupNotifications();
}

export function IsNotificationAvailable() {
    return _appWindow.runtime.IsNotificationAvailable();
}

export function RequestNotificationAuthorization() {
    return _appWindow.runtime.RequestNotificationAuthorization();
}

export function CheckNotificationAuthorization() {
    return _appWindow.runtime.CheckNotificationAuthorization();
}

export function SendNotification(options) {
    return _appWindow.runtime.SendNotification(options);
}

export function SendNotificationWithActions(options) {
    return _appWindow.runtime.SendNotificationWithActions(options);
}

export function RegisterNotificationCategory(category) {
    return _appWindow.runtime.RegisterNotificationCategory(category);
}

export function RemoveNotificationCategory(categoryId) {
    return _appWindow.runtime.RemoveNotificationCategory(categoryId);
}

export function RemoveAllPendingNotifications() {
    return _appWindow.runtime.RemoveAllPendingNotifications();
}

export function RemovePendingNotification(identifier) {
    return _appWindow.runtime.RemovePendingNotification(identifier);
}

export function RemoveAllDeliveredNotifications() {
    return _appWindow.runtime.RemoveAllDeliveredNotifications();
}

export function RemoveDeliveredNotification(identifier) {
    return _appWindow.runtime.RemoveDeliveredNotification(identifier);
}

export function RemoveNotification(identifier) {
    return _appWindow.runtime.RemoveNotification(identifier);
}