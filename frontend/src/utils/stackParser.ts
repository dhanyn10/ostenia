/**
 * Information about the caller of a log function.
 */
export interface CallerInfo {
  functionName: string;
  fileName: string;
  line: string;
  column: string;
}

/**
 * Result of parsing a stack trace.
 */
export interface ParsedStack {
  caller?: CallerInfo;
  rawStack: string;
}

/**
 * Cleans a file path to extract only the base file name.
 * For example: "http://localhost:5173/src/App.tsx?t=12345" -> "App.tsx".
 *
 * @param path The raw file path or URL from the stack trace line.
 * @returns The cleaned file name.
 */
export function cleanFileName(path: string): string {
  if (!path) return '';
  // Remove query parameters or hash fragments
  let clean = path.split('?')[0].split('#')[0];
  // Get base name
  const lastSlash = Math.max(clean.lastIndexOf('/'), clean.lastIndexOf('\\'));
  if (lastSlash !== -1) {
    clean = clean.substring(lastSlash + 1);
  }
  return clean;
}

/**
 * Parses a raw JS/TS error stack trace to extract the original caller's location.
 * It filters out internal logging frameworks, console overrides, and utility frames.
 *
 * @param stack The raw stack trace string from `new Error().stack`.
 * @returns An object containing the parsed caller info and the raw stack trace.
 */
export function parseStackTrace(stack: string | undefined): ParsedStack {
  const rawStack = stack || '';
  if (!rawStack) {
    return { rawStack };
  }

  const lines = rawStack.split('\n');
  const frames: CallerInfo[] = [];

  for (let line of lines) {
    line = line.trim();
    if (!line) continue;

    // V8 format with function name: "at functionName (filePath:line:column)"
    const v8WithFunc = line.match(/^at\s+([^(]+)\s+\((.+):(\d+):(\d+)\)$/);
    if (v8WithFunc) {
      frames.push({
        functionName: v8WithFunc[1].trim(),
        fileName: cleanFileName(v8WithFunc[2]),
        line: v8WithFunc[3],
        column: v8WithFunc[4],
      });
      continue;
    }

    // V8 format without function name: "at filePath:line:column"
    const v8NoFunc = line.match(/^at\s+(.+):(\d+):(\d+)$/);
    if (v8NoFunc) {
      frames.push({
        functionName: 'anonymous',
        fileName: cleanFileName(v8NoFunc[1]),
        line: v8NoFunc[2],
        column: v8NoFunc[3],
      });
      continue;
    }

    // Firefox/Safari format with function name: "functionName@filePath:line:column"
    const ffWithFunc = line.match(/^([^@]+)@(.+):(\d+):(\d+)$/);
    if (ffWithFunc) {
      frames.push({
        functionName: ffWithFunc[1].trim(),
        fileName: cleanFileName(ffWithFunc[2]),
        line: ffWithFunc[3],
        column: ffWithFunc[4],
      });
      continue;
    }

    // Firefox/Safari format without function name: "@filePath:line:column"
    const ffNoFunc = line.match(/^@(.+):(\d+):(\d+)$/);
    if (ffNoFunc) {
      frames.push({
        functionName: 'anonymous',
        fileName: cleanFileName(ffNoFunc[1]),
        line: ffNoFunc[2],
        column: ffNoFunc[3],
      });
      continue;
    }
  }

  // List of patterns to skip so we find the actual caller function
  const ignorePatterns = [
    /addLog/i,
    /setupConsoleOverrides/i,
    /stackParser/i,
    /parseStackTrace/i,
    /console\.(log|warn|error|info)/i,
    /Object\.(log|warn|error|info)/i,
    /at\s+log\s+\(/i,
    /at\s+warn\s+\(/i,
    /at\s+error\s+\(/i,
  ];

  const caller = frames.find(frame => {
    const isIgnored = ignorePatterns.some(pattern =>
      pattern.test(frame.functionName) || pattern.test(frame.fileName)
    );
    return !isIgnored;
  });

  return {
    caller,
    rawStack,
  };
}
