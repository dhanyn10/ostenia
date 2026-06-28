const appWindow = appWindow; // NOSONAR
/*
 _       __      _ __
| |     / /___ _(_) /____
| | /| / / __ `/ / / ___/
| |/ |/ / /_/ / / (__  )
|__/|__/\__,_/_/_/____/
The electron alternative for Go
(c) Lea Anthony 2019-present
*/

export function LogPrint(message) {
    appWindow.runtime.LogPrint(message);
}

export function LogTrace(message) {
    appWindow.runtime.LogTrace(message);
}

export function LogDebug(message) {
    appWindow.runtime.LogDebug(message);
}

export function LogInfo(message) {
    appWindow.runtime.LogInfo(message);
}

export function LogWarning(message) {
    appWindow.runtime.LogWarning(message);
}

export function LogError(message) {
    appWindow.runtime.LogError(message);
}

export function LogFatal(message) {
    appWindow.runtime.LogFatal(message);
}

export function EventsOnMultiple(eventName, callback, maxCallbacks) {
    return appWindow.runtime.EventsOnMultiple(eventName, callback, maxCallbacks);
}

export function EventsOn(eventName, callback) {
    return EventsOnMultiple(eventName, callback, -1);
}

export function EventsOff(eventName, ...additionalEventNames) {
    return appWindow.runtime.EventsOff(eventName, ...additionalEventNames);
}

export function EventsOffAll() {
  return appWindow.runtime.EventsOffAll();
}

export function EventsOnce(eventName, callback) {
    return EventsOnMultiple(eventName, callback, 1);
}

export function EventsEmit(eventName) {
    let args = [eventName].slice.call(arguments);
    return appWindow.runtime.EventsEmit.apply(null, args);
}

export function WindowReload() {
    appWindow.runtime.WindowReload();
}

export function WindowReloadApp() {
    appWindow.runtime.WindowReloadApp();
}

export function WindowSetAlwaysOnTop(b) {
    appWindow.runtime.WindowSetAlwaysOnTop(b);
}

export function WindowSetSystemDefaultTheme() {
    appWindow.runtime.WindowSetSystemDefaultTheme();
}

export function WindowSetLightTheme() {
    appWindow.runtime.WindowSetLightTheme();
}

export function WindowSetDarkTheme() {
    appWindow.runtime.WindowSetDarkTheme();
}

export function WindowCenter() {
    appWindow.runtime.WindowCenter();
}

export function WindowSetTitle(title) {
    appWindow.runtime.WindowSetTitle(title);
}

export function WindowFullscreen() {
    appWindow.runtime.WindowFullscreen();
}

export function WindowUnfullscreen() {
    appWindow.runtime.WindowUnfullscreen();
}

export function WindowIsFullscreen() {
    return appWindow.runtime.WindowIsFullscreen();
}

export function WindowGetSize() {
    return appWindow.runtime.WindowGetSize();
}

export function WindowSetSize(width, height) {
    appWindow.runtime.WindowSetSize(width, height);
}

export function WindowSetMaxSize(width, height) {
    appWindow.runtime.WindowSetMaxSize(width, height);
}

export function WindowSetMinSize(width, height) {
    appWindow.runtime.WindowSetMinSize(width, height);
}

export function WindowSetPosition(x, y) {
    appWindow.runtime.WindowSetPosition(x, y);
}

export function WindowGetPosition() {
    return appWindow.runtime.WindowGetPosition();
}

export function WindowHide() {
    appWindow.runtime.WindowHide();
}

export function WindowShow() {
    appWindow.runtime.WindowShow();
}

export function WindowMaximise() {
    appWindow.runtime.WindowMaximise();
}

export function WindowToggleMaximise() {
    appWindow.runtime.WindowToggleMaximise();
}

export function WindowUnmaximise() {
    appWindow.runtime.WindowUnmaximise();
}

export function WindowIsMaximised() {
    return appWindow.runtime.WindowIsMaximised();
}

export function WindowMinimise() {
    appWindow.runtime.WindowMinimise();
}

export function WindowUnminimise() {
    appWindow.runtime.WindowUnminimise();
}

export function WindowSetBackgroundColour(R, G, B, A) {
    appWindow.runtime.WindowSetBackgroundColour(R, G, B, A);
}

export function ScreenGetAll() {
    return appWindow.runtime.ScreenGetAll();
}

export function WindowIsMinimised() {
    return appWindow.runtime.WindowIsMinimised();
}

export function WindowIsNormal() {
    return appWindow.runtime.WindowIsNormal();
}

export function BrowserOpenURL(url) {
    appWindow.runtime.BrowserOpenURL(url);
}

export function Environment() {
    return appWindow.runtime.Environment();
}

export function Quit() {
    appWindow.runtime.Quit();
}

export function Hide() {
    appWindow.runtime.Hide();
}

export function Show() {
    appWindow.runtime.Show();
}

export function ClipboardGetText() {
    return appWindow.runtime.ClipboardGetText();
}

export function ClipboardSetText(text) {
    return appWindow.runtime.ClipboardSetText(text);
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
    return appWindow.runtime.OnFileDrop(callback, useDropTarget);
}

/**
 * OnFileDropOff removes the drag and drop listeners and handlers.
 */
export function OnFileDropOff() {
    return appWindow.runtime.OnFileDropOff();
}

export function CanResolveFilePaths() {
    return appWindow.runtime.CanResolveFilePaths();
}

export function ResolveFilePaths(files) {
    return appWindow.runtime.ResolveFilePaths(files);
}

export function InitializeNotifications() {
    return appWindow.runtime.InitializeNotifications();
}

export function CleanupNotifications() {
    return appWindow.runtime.CleanupNotifications();
}

export function IsNotificationAvailable() {
    return appWindow.runtime.IsNotificationAvailable();
}

export function RequestNotificationAuthorization() {
    return appWindow.runtime.RequestNotificationAuthorization();
}

export function CheckNotificationAuthorization() {
    return appWindow.runtime.CheckNotificationAuthorization();
}

export function SendNotification(options) {
    return appWindow.runtime.SendNotification(options);
}

export function SendNotificationWithActions(options) {
    return appWindow.runtime.SendNotificationWithActions(options);
}

export function RegisterNotificationCategory(category) {
    return appWindow.runtime.RegisterNotificationCategory(category);
}

export function RemoveNotificationCategory(categoryId) {
    return appWindow.runtime.RemoveNotificationCategory(categoryId);
}

export function RemoveAllPendingNotifications() {
    return appWindow.runtime.RemoveAllPendingNotifications();
}

export function RemovePendingNotification(identifier) {
    return appWindow.runtime.RemovePendingNotification(identifier);
}

export function RemoveAllDeliveredNotifications() {
    return appWindow.runtime.RemoveAllDeliveredNotifications();
}

export function RemoveDeliveredNotification(identifier) {
    return appWindow.runtime.RemoveDeliveredNotification(identifier);
}

export function RemoveNotification(identifier) {
    return appWindow.runtime.RemoveNotification(identifier);
}