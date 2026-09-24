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

    let functionName = '';
    let location = '';

    if (line.startsWith('at ')) {
      const lastParen = line.lastIndexOf('(');
      if (line.endsWith(')') && lastParen !== -1) {
        functionName = line.substring(3, lastParen).trim();
        location = line.substring(lastParen + 1, line.length - 1);
      } else {
        functionName = 'anonymous';
        location = line.substring(3).trim();
      }
    } else if (line.includes('@')) {
      const atIdx = line.indexOf('@');
      functionName = line.substring(0, atIdx).trim() || 'anonymous';
      location = line.substring(atIdx + 1);
    } else {
      continue;
    }

    const lastColon = location.lastIndexOf(':');
    const secondLastColon = location.lastIndexOf(':', lastColon - 1);
    if (lastColon === -1 || secondLastColon === -1) continue;

    const lineNum = location.substring(secondLastColon + 1, lastColon);
    const colNum = location.substring(lastColon + 1);

    if (!/^\d+$/.test(lineNum) || !/^\d+$/.test(colNum)) continue;

    const rawFilePath = location.substring(0, secondLastColon);

    frames.push({
      functionName: functionName || 'anonymous',
      fileName: cleanFileName(rawFilePath),
      line: lineNum,
      column: colNum,
    });
  }

  // List of patterns to skip so we find the actual caller function
  const ignorePatterns = [
    /addLog/i,
    /setupConsoleOverrides/i,
    /stackParser/i,
    /parseStackTrace/i,
    /measureActivity/i,
    /activityLogger/i,
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
