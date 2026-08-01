export interface CallerInfo {
  functionName: string;
  fileName: string;
  line: string;
  column: string;
  rawStack: string;
}

export function getCallerInfo(): CallerInfo {
  const err = new Error();
  const rawStack = err.stack || "";
  const lines = rawStack.split("\n");

  let targetLine = "";
  // Skip frames that are part of the logging utility itself
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    if (!line || line.trim() === "") continue;
    if (line.includes("Error") && i === 0) continue; // Skip first line if it's just "Error"

    // Skip utility functions
    if (
      line.includes("getCallerInfo") ||
      line.includes("stackParser") ||
      line.includes("addLog") ||
      line.includes("setupConsoleOverrides") ||
      line.includes("console.log") ||
      line.includes("console.warn") ||
      line.includes("console.error")
    ) {
      continue;
    }
    targetLine = line;
    break;
  }

  if (!targetLine) {
    return {
      functionName: "unknown",
      fileName: "unknown",
      line: "0",
      column: "0",
      rawStack,
    };
  }

  // Parse targetLine
  // Format 1: "    at functionName (file:line:col)"
  // Format 2: "    at file:line:col"
  // Format 3: "functionName@file:line:col"
  // Format 4: "file:line:col"

  let functionName = "anonymous";
  let locationPart = targetLine.trim();

  // Try parsing Firefox/Safari "@" format first
  if (locationPart.includes("@")) {
    const atParts = locationPart.split("@");
    functionName = atParts[0] || "anonymous";
    locationPart = atParts[1] || "";
  } else if (locationPart.startsWith("at ")) {
    // V8 format "at ..."
    const content = locationPart.slice(3).trim();
    const openParen = content.indexOf("(");
    const closeParen = content.lastIndexOf(")");
    if (openParen !== -1 && closeParen !== -1) {
      functionName = content.slice(0, openParen).trim();
      locationPart = content.slice(openParen + 1, closeParen).trim();
    } else {
      locationPart = content;
    }
  }

  // Clean location part from query params (e.g. ?t=12312412)
  // Usually file_path?params:line:col
  let cleanLocation = locationPart;
  const qIdx = cleanLocation.indexOf("?");
  if (qIdx !== -1) {
    const nextColon = cleanLocation.indexOf(":", qIdx);
    if (nextColon !== -1) {
      cleanLocation = cleanLocation.slice(0, qIdx) + cleanLocation.slice(nextColon);
    } else {
      cleanLocation = cleanLocation.slice(0, qIdx);
    }
  }

  // Now split cleanLocation by ":" to get fileName, line, col
  const parts = cleanLocation.split(":");
  let fileName = "unknown";
  let line = "0";
  let column = "0";

  if (parts.length >= 3) {
    column = parts[parts.length - 1];
    line = parts[parts.length - 2];
    const fileParts = parts.slice(0, parts.length - 2);
    const fullPath = fileParts.join(":");
    // Get last segment of path
    const slashIdx = fullPath.lastIndexOf("/");
    fileName = slashIdx !== -1 ? fullPath.slice(slashIdx + 1) : fullPath;
  } else if (parts.length === 2) {
    line = parts[1];
    const fullPath = parts[0];
    const slashIdx = fullPath.lastIndexOf("/");
    fileName = slashIdx !== -1 ? fullPath.slice(slashIdx + 1) : fullPath;
  } else {
    const slashIdx = cleanLocation.lastIndexOf("/");
    fileName = slashIdx !== -1 ? cleanLocation.slice(slashIdx + 1) : cleanLocation;
  }

  return {
    functionName: functionName || "anonymous",
    fileName: fileName || "unknown",
    line: line || "0",
    column: column || "0",
    rawStack,
  };
}
