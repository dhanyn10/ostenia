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
    window.runtime.LogPrint(message); // NOSONAR
}

export function LogTrace(message) {
    window.runtime.LogTrace(message); // NOSONAR
}

export function LogDebug(message) {
    window.runtime.LogDebug(message); // NOSONAR
}

export function LogInfo(message) {
    window.runtime.LogInfo(message); // NOSONAR
}

export function LogWarning(message) {
    window.runtime.LogWarning(message); // NOSONAR
}

export function LogError(message) {
    window.runtime.LogError(message); // NOSONAR
}

export function LogFatal(message) {
    window.runtime.LogFatal(message); // NOSONAR
}

export function EventsOnMultiple(eventName, callback, maxCallbacks) {
    return window.runtime.EventsOnMultiple(eventName, callback, maxCallbacks); // NOSONAR
}

export function EventsOn(eventName, callback) {
    return EventsOnMultiple(eventName, callback, -1);
}

export function EventsOff(eventName, ...additionalEventNames) {
    return window.runtime.EventsOff(eventName, ...additionalEventNames); // NOSONAR
}

export function EventsOffAll() {
  return window.runtime.EventsOffAll(); // NOSONAR
}

export function EventsOnce(eventName, callback) {
    return EventsOnMultiple(eventName, callback, 1);
}

export function EventsEmit(eventName) {
    let args = [eventName].slice.call(arguments);
    return window.runtime.EventsEmit.apply(null, args); // NOSONAR
}

export function WindowReload() {
    window.runtime.WindowReload(); // NOSONAR
}

export function WindowReloadApp() {
    window.runtime.WindowReloadApp(); // NOSONAR
}

export function WindowSetAlwaysOnTop(b) {
    window.runtime.WindowSetAlwaysOnTop(b); // NOSONAR
}

export function WindowSetSystemDefaultTheme() {
    window.runtime.WindowSetSystemDefaultTheme(); // NOSONAR
}

export function WindowSetLightTheme() {
    window.runtime.WindowSetLightTheme(); // NOSONAR
}

export function WindowSetDarkTheme() {
    window.runtime.WindowSetDarkTheme(); // NOSONAR
}

export function WindowCenter() {
    window.runtime.WindowCenter(); // NOSONAR
}

export function WindowSetTitle(title) {
    window.runtime.WindowSetTitle(title); // NOSONAR
}

export function WindowFullscreen() {
    window.runtime.WindowFullscreen(); // NOSONAR
}

export function WindowUnfullscreen() {
    window.runtime.WindowUnfullscreen(); // NOSONAR
}

export function WindowIsFullscreen() {
    return window.runtime.WindowIsFullscreen(); // NOSONAR
}

export function WindowGetSize() {
    return window.runtime.WindowGetSize(); // NOSONAR
}

export function WindowSetSize(width, height) {
    window.runtime.WindowSetSize(width, height); // NOSONAR
}

export function WindowSetMaxSize(width, height) {
    window.runtime.WindowSetMaxSize(width, height); // NOSONAR
}

export function WindowSetMinSize(width, height) {
    window.runtime.WindowSetMinSize(width, height); // NOSONAR
}

export function WindowSetPosition(x, y) {
    window.runtime.WindowSetPosition(x, y); // NOSONAR
}

export function WindowGetPosition() {
    return window.runtime.WindowGetPosition(); // NOSONAR
}

export function WindowHide() {
    window.runtime.WindowHide(); // NOSONAR
}

export function WindowShow() {
    window.runtime.WindowShow(); // NOSONAR
}

export function WindowMaximise() {
    window.runtime.WindowMaximise(); // NOSONAR
}

export function WindowToggleMaximise() {
    window.runtime.WindowToggleMaximise(); // NOSONAR
}

export function WindowUnmaximise() {
    window.runtime.WindowUnmaximise(); // NOSONAR
}

export function WindowIsMaximised() {
    return window.runtime.WindowIsMaximised(); // NOSONAR
}

export function WindowMinimise() {
    window.runtime.WindowMinimise(); // NOSONAR
}

export function WindowUnminimise() {
    window.runtime.WindowUnminimise(); // NOSONAR
}

export function WindowSetBackgroundColour(R, G, B, A) {
    window.runtime.WindowSetBackgroundColour(R, G, B, A); // NOSONAR
}

export function ScreenGetAll() {
    return window.runtime.ScreenGetAll(); // NOSONAR
}

export function WindowIsMinimised() {
    return window.runtime.WindowIsMinimised(); // NOSONAR
}

export function WindowIsNormal() {
    return window.runtime.WindowIsNormal(); // NOSONAR
}

export function BrowserOpenURL(url) {
    window.runtime.BrowserOpenURL(url); // NOSONAR
}

export function Environment() {
    return window.runtime.Environment(); // NOSONAR
}

export function Quit() {
    window.runtime.Quit(); // NOSONAR
}

export function Hide() {
    window.runtime.Hide(); // NOSONAR
}

export function Show() {
    window.runtime.Show(); // NOSONAR
}

export function ClipboardGetText() {
    return window.runtime.ClipboardGetText(); // NOSONAR
}

export function ClipboardSetText(text) {
    return window.runtime.ClipboardSetText(text); // NOSONAR
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
    return window.runtime.OnFileDrop(callback, useDropTarget); // NOSONAR
}

/**
 * OnFileDropOff removes the drag and drop listeners and handlers.
 */
export function OnFileDropOff() {
    return window.runtime.OnFileDropOff(); // NOSONAR
}

export function CanResolveFilePaths() {
    return window.runtime.CanResolveFilePaths(); // NOSONAR
}

export function ResolveFilePaths(files) {
    return window.runtime.ResolveFilePaths(files); // NOSONAR
}

export function InitializeNotifications() {
    return window.runtime.InitializeNotifications(); // NOSONAR
}

export function CleanupNotifications() {
    return window.runtime.CleanupNotifications(); // NOSONAR
}

export function IsNotificationAvailable() {
    return window.runtime.IsNotificationAvailable(); // NOSONAR
}

export function RequestNotificationAuthorization() {
    return window.runtime.RequestNotificationAuthorization(); // NOSONAR
}

export function CheckNotificationAuthorization() {
    return window.runtime.CheckNotificationAuthorization(); // NOSONAR
}

export function SendNotification(options) {
    return window.runtime.SendNotification(options); // NOSONAR
}

export function SendNotificationWithActions(options) {
    return window.runtime.SendNotificationWithActions(options); // NOSONAR
}

export function RegisterNotificationCategory(category) {
    return window.runtime.RegisterNotificationCategory(category); // NOSONAR
}

export function RemoveNotificationCategory(categoryId) {
    return window.runtime.RemoveNotificationCategory(categoryId); // NOSONAR
}

export function RemoveAllPendingNotifications() {
    return window.runtime.RemoveAllPendingNotifications(); // NOSONAR
}

export function RemovePendingNotification(identifier) {
    return window.runtime.RemovePendingNotification(identifier); // NOSONAR
}

export function RemoveAllDeliveredNotifications() {
    return window.runtime.RemoveAllDeliveredNotifications(); // NOSONAR
}

export function RemoveDeliveredNotification(identifier) {
    return window.runtime.RemoveDeliveredNotification(identifier); // NOSONAR
}

export function RemoveNotification(identifier) {
    return window.runtime.RemoveNotification(identifier); // NOSONAR
}