import type { LogEntry, LogLevel, PaginationItem } from "@/types/logs";

export function getPaginationRange(
  currentPage: number,
  totalPages: number,
  siblingCount = 1,
): PaginationItem[] {
  const totalPageNumbers = siblingCount + 5; // siblingCount + firstPage + lastPage + currentPage + 2*ellipses

  if (totalPageNumbers >= totalPages) {
    return Array.from({ length: totalPages }, (_, i) => ({
      key: `page-${i + 1}`,
      value: i + 1,
      isDots: false,
    }));
  }

  const leftSiblingIndex = Math.max(currentPage - siblingCount, 1);
  const rightSiblingIndex = Math.min(currentPage + siblingCount, totalPages);

  const shouldShowLeftDots = leftSiblingIndex > 2;
  const shouldShowRightDots = rightSiblingIndex < totalPages - 2;

  if (!shouldShowLeftDots && shouldShowRightDots) {
    const leftItemCount = 3 + 2 * siblingCount;
    const leftRange = Array.from({ length: leftItemCount }, (_, i) => ({
      key: `page-${i + 1}`,
      value: i + 1,
      isDots: false,
    }));
    return [
      ...leftRange,
      { key: "dots-right", value: "...", isDots: true },
      { key: `page-${totalPages}`, value: totalPages, isDots: false },
    ];
  }

  if (shouldShowLeftDots && !shouldShowRightDots) {
    const rightItemCount = 3 + 2 * siblingCount;
    const rightRange = Array.from({ length: rightItemCount }, (_, i) => {
      const pageNum = totalPages - rightItemCount + i + 1;
      return { key: `page-${pageNum}`, value: pageNum, isDots: false };
    });
    return [
      { key: "page-1", value: 1, isDots: false },
      { key: "dots-left", value: "...", isDots: true },
      ...rightRange,
    ];
  }

  if (shouldShowLeftDots && shouldShowRightDots) {
    const middleRange = Array.from(
      { length: rightSiblingIndex - leftSiblingIndex + 1 },
      (_, i) => {
        const pageNum = leftSiblingIndex + i;
        return { key: `page-${pageNum}`, value: pageNum, isDots: false };
      },
    );
    return [
      { key: "page-1", value: 1, isDots: false },
      { key: "dots-left", value: "...", isDots: true },
      ...middleRange,
      { key: "dots-right", value: "...", isDots: true },
      { key: `page-${totalPages}`, value: totalPages, isDots: false },
    ];
  }

  return [];
}

export function getLogLevel(log: LogEntry): LogLevel {
  if (log.isServiceLog) return "SERVICE";
  const type = (log.type || "").toLowerCase();
  if (type === "error" || type === "err") return "ERROR";
  if (type === "warn" || type === "wrn") return "WARN";

  // Fallback to checking the msg
  const msg = (log.msg || "").toLowerCase();
  if (
    msg.includes("[err]") ||
    msg.includes("error") ||
    msg.includes("failed") ||
    msg.includes("crash") ||
    msg.includes("fatal") ||
    msg.includes("panic") ||
    msg.includes("exception")
  ) {
    return "ERROR";
  }
  if (msg.includes("[wrn]")) return "WARN";
  return "INFO";
}

export function formatCompleteLog(log: LogEntry): string {
  const level = getLogLevel(log);
  const timeStr = log.time || "";

  let message = log.cleanMsg || log.msg || "";
  if (
    message.startsWith("[SYS] ") ||
    message.startsWith("[ERR] ") ||
    message.startsWith("[WRN] ")
  ) {
    message = message.substring(6);
  }

  const callerFunc = log.caller ? `${log.caller.functionName}()` : "N/A";
  const fileInfo = log.caller
    ? `${log.caller.fileName}:${log.caller.line}:${log.caller.column}`
    : "N/A";
  const stackTraceText = log.rawStack ? log.rawStack : "N/A";

  return [
    `Timestamp: [${timeStr}]`,
    `Level: ${level}`,
    `Message: ${message}`,
    `Caller Function: ${callerFunc}`,
    `File: ${fileInfo}`,
    `Stack Trace:\n${stackTraceText}`,
  ].join("\n");
}
