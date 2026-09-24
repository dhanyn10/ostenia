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
 * Parses a single raw line from a stack trace into a CallerInfo object.
 *
 * @param line Single raw stack line.
 * @returns Parsed CallerInfo or null if unparseable.
 */
export function parseStackLine(line: string): CallerInfo | null {
  const trimmed = line.trim();
  if (!trimmed) return null;

  let functionName = '';
  let location = '';

  if (trimmed.startsWith('at ')) {
    const lastParen = trimmed.lastIndexOf('(');
    if (trimmed.endsWith(')') && lastParen !== -1) {
      functionName = trimmed.substring(3, lastParen).trim();
      location = trimmed.substring(lastParen + 1, trimmed.length - 1);
    } else {
      functionName = 'anonymous';
      location = trimmed.substring(3).trim();
    }
  } else if (trimmed.includes('@')) {
    const atIdx = trimmed.indexOf('@');
    functionName = trimmed.substring(0, atIdx).trim() || 'anonymous';
    location = trimmed.substring(atIdx + 1);
  } else {
    return null;
  }

  const lastColon = location.lastIndexOf(':');
  const secondLastColon = location.lastIndexOf(':', lastColon - 1);
  if (lastColon === -1 || secondLastColon === -1) return null;

  const lineNum = location.substring(secondLastColon + 1, lastColon);
  const colNum = location.substring(lastColon + 1);

  if (!/^\d+$/.test(lineNum) || !/^\d+$/.test(colNum)) return null;

  const rawFilePath = location.substring(0, secondLastColon);

  return {
    functionName: functionName || 'anonymous',
    fileName: cleanFileName(rawFilePath),
    line: lineNum,
    column: colNum,
  };
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

function isIgnoredFrame(frame: CallerInfo): boolean {
  return ignorePatterns.some(pattern =>
    pattern.test(frame.functionName) || pattern.test(frame.fileName)
  );
}

/**
 * Parses a raw JS/TS error stack trace to extract the original caller's location.
 * It filters out internal logging frameworks, console overrides, and utility frames.
 *
 * @param stack The raw stack trace string from `new Error().stack`.
 * @returns An object containing the parsed caller info and the raw stack trace.
 */
export function parseStackTrace(stack: string = ''): ParsedStack {
  if (!stack) {
    return { rawStack: stack };
  }
  const rawStack = stack;

  const lines = rawStack.split('\n');
  const frames: CallerInfo[] = [];

  for (const line of lines) {
    const frame = parseStackLine(line);
    if (frame) {
      frames.push(frame);
    }
  }

  const caller = frames.find(frame => !isIgnoredFrame(frame));

  return {
    caller,
    rawStack,
  };
}
