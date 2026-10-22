import { describe, it, expect } from "vitest";
import {
  getLogLevel,
  formatCompleteLog,
  getPaginationRange,
} from "@/utils/logFormatters";
import type { LogEntry } from "@/types/logs";

describe("logFormatters utility", () => {
  describe("getLogLevel", () => {
    it("returns SERVICE if isServiceLog is true", () => {
      const log: LogEntry = {
        id: 1,
        time: "10:00:00",
        msg: "service starting",
        isServiceLog: true,
      };
      expect(getLogLevel(log)).toBe("SERVICE");
    });

    it("returns ERROR if type is error or err", () => {
      expect(
        getLogLevel({ id: 1, time: "10:00", msg: "some msg", type: "error" }),
      ).toBe("ERROR");
      expect(
        getLogLevel({ id: 2, time: "10:00", msg: "some msg", type: "err" }),
      ).toBe("ERROR");
    });

    it("returns WARN if type is warn or wrn", () => {
      expect(
        getLogLevel({ id: 1, time: "10:00", msg: "some msg", type: "warn" }),
      ).toBe("WARN");
      expect(
        getLogLevel({ id: 2, time: "10:00", msg: "some msg", type: "wrn" }),
      ).toBe("WARN");
    });

    it("identifies ERROR from msg keywords", () => {
      expect(
        getLogLevel({ id: 1, time: "10:00", msg: "Fatal database crash" }),
      ).toBe("ERROR");
      expect(
        getLogLevel({ id: 2, time: "10:00", msg: "Unhandled exception" }),
      ).toBe("ERROR");
      expect(
        getLogLevel({ id: 3, time: "10:00", msg: "Service panic happened" }),
      ).toBe("ERROR");
      expect(
        getLogLevel({ id: 4, time: "10:00", msg: "Connection failed" }),
      ).toBe("ERROR");
      expect(
        getLogLevel({ id: 5, time: "10:00", msg: "[ERR] something broke" }),
      ).toBe("ERROR");
    });

    it("identifies WARN from [wrn] tag in msg", () => {
      expect(
        getLogLevel({ id: 1, time: "10:00", msg: "[wrn] high memory usage" }),
      ).toBe("WARN");
    });

    it("defaults to INFO if no error/warn indicators", () => {
      expect(
        getLogLevel({ id: 1, time: "10:00", msg: "Server started on port 80" }),
      ).toBe("INFO");
    });
  });

  describe("formatCompleteLog", () => {
    it("formats log entry with all fields present", () => {
      const log: LogEntry = {
        id: 1,
        time: "10:15:30",
        msg: "[SYS] System initialized",
        caller: {
          functionName: "initApp",
          fileName: "App.tsx",
          line: "42",
          column: "5",
        },
        rawStack: "Error at initApp (App.tsx:42:5)",
      };

      const formatted = formatCompleteLog(log);
      expect(formatted).toContain("Timestamp: [10:15:30]");
      expect(formatted).toContain("Level: INFO");
      expect(formatted).toContain("Message: System initialized");
      expect(formatted).toContain("Caller Function: initApp()");
      expect(formatted).toContain("File: App.tsx:42:5");
      expect(formatted).toContain(
        "Stack Trace:\nError at initApp (App.tsx:42:5)",
      );
    });

    it("formats log entry with missing caller and rawStack as N/A", () => {
      const log: LogEntry = {
        id: 2,
        time: "12:00:00",
        msg: "Simple message",
      };

      const formatted = formatCompleteLog(log);
      expect(formatted).toContain("Caller Function: N/A");
      expect(formatted).toContain("File: N/A");
      expect(formatted).toContain("Stack Trace:\nN/A");
    });
  });

  describe("getPaginationRange", () => {
    it("returns simple page list when totalPages is less than totalPageNumbers", () => {
      const range = getPaginationRange(1, 4, 1);
      expect(range).toHaveLength(4);
      expect(range.map((r) => r.value)).toEqual([1, 2, 3, 4]);
    });

    it("shows right dots when currentPage is near start", () => {
      const range = getPaginationRange(1, 10, 1);
      expect(range.some((r) => r.key === "dots-right")).toBe(true);
      expect(range.at(-1)?.value).toBe(10);
    });

    it("shows left dots when currentPage is near end", () => {
      const range = getPaginationRange(10, 10, 1);
      expect(range.some((r) => r.key === "dots-left")).toBe(true);
      expect(range[0].value).toBe(1);
    });

    it("shows both left and right dots when currentPage is in the middle", () => {
      const range = getPaginationRange(5, 10, 1);
      expect(range.some((r) => r.key === "dots-left")).toBe(true);
      expect(range.some((r) => r.key === "dots-right")).toBe(true);
      expect(range[0].value).toBe(1);
      expect(range.at(-1)?.value).toBe(10);
    });
  });
});
