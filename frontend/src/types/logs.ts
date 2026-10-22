export interface CallerInfo {
  functionName: string;
  fileName: string;
  line: string;
  column: string;
}

export type LogId = string | number;
export type LogLevel = "SERVICE" | "ERROR" | "WARN" | "INFO";
export type LogViewMode = "simple" | "complete";

export interface LogEntry {
  id: LogId;
  time: string;
  timestamp?: string;
  msg: string;
  cleanMsg?: string;
  type?: string;
  isServiceLog?: boolean;
  caller?: CallerInfo;
  rawStack?: string;
}

export type LogItem = LogEntry;

export interface PaginationItem {
  key: string;
  value: number | string;
  isDots: boolean;
}
