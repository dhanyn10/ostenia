import { describe, it, expect } from "vitest";
import { cleanFileName, parseStackTrace } from "./stackParser";

describe("stackParser Utility Table-Driven Tests", () => {
  describe("cleanFileName", () => {
    const filenameCases = [
      { input: "App.tsx", expected: "App.tsx" },
      { input: "http://localhost:5173/src/App.tsx", expected: "App.tsx" },
      { input: "http://localhost:5173/src/App.tsx?t=123456789", expected: "App.tsx" },
      { input: "http://localhost:5173/src/App.tsx#section", expected: "App.tsx" },
      { input: "C:\\projects\\ostenia\\src\\main.tsx", expected: "main.tsx" },
      { input: "/home/user/ostenia/src/main.tsx", expected: "main.tsx" },
      { input: "", expected: "" },
    ];

    it.each(filenameCases)("should clean filename from %s to %s", ({ input, expected }) => {
      expect(cleanFileName(input)).toBe(expected);
    });
  });

  describe("parseStackTrace", () => {
    it("handles empty stack trace", () => {
      const res = parseStackTrace("");
      expect(res.rawStack).toBe("");
      expect(res.caller).toBeUndefined();
    });

    const stackTraceCases = [
      {
        name: "V8 stack with function name",
        stack: `Error
    at addLog (http://localhost:5173/src/App.tsx:130:15)
    at handleToggleService (http://localhost:5173/src/App.tsx:392:5)
    at HTMLButtonElement.onClick (http://localhost:5173/src/components/ServiceItem.tsx:20:10)`,
        expected: { functionName: "handleToggleService", fileName: "App.tsx", line: "392", column: "5" }
      },
      {
        name: "V8 stack without function name (anonymous)",
        stack: `Error
    at http://localhost:5173/src/App.tsx:392:5`,
        expected: { functionName: "anonymous", fileName: "App.tsx", line: "392", column: "5" }
      },
      {
        name: "Firefox/Safari stack format",
        stack: `addLog@http://localhost:5173/src/App.tsx:130:15
handleToggleService@http://localhost:5173/src/App.tsx:392:5
onClick@http://localhost:5173/src/components/ServiceItem.tsx:20:10`,
        expected: { functionName: "handleToggleService", fileName: "App.tsx", line: "392", column: "5" }
      },
      {
        name: "ignore list filtering of logging wrapper frames",
        stack: `Error
    at parseStackTrace (http://localhost:5173/src/utils/stackParser.ts:35:10)
    at addLog (http://localhost:5173/src/App.tsx:130:15)
    at Object.console.log (http://localhost:5173/src/App.tsx:245:5)
    at myCoolAppFunc (http://localhost:5173/src/components/SettingsModal.tsx:40:12)`,
        expected: { functionName: "myCoolAppFunc", fileName: "SettingsModal.tsx", line: "40", column: "12" }
      }
    ];

    it.each(stackTraceCases)("parses stack for: %s", ({ stack, expected }) => {
      const result = parseStackTrace(stack);
      expect(result.caller).toBeDefined();
      expect(result.caller?.functionName).toBe(expected.functionName);
      expect(result.caller?.fileName).toBe(expected.fileName);
      if (expected.line) {
        expect(result.caller?.line).toBe(expected.line);
        expect(result.caller?.column).toBe(expected.column);
      }
    });
  });
});
