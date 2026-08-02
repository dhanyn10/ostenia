import { describe, it, expect } from "vitest";
import { cleanFileName, parseStackTrace } from "./stackParser";

describe("stackParser Utility", () => {
  describe("cleanFileName", () => {
    it("handles standard files", () => {
      expect(cleanFileName("App.tsx")).toBe("App.tsx");
    });

    it("handles URL file paths", () => {
      expect(cleanFileName("http://localhost:5173/src/App.tsx")).toBe("App.tsx");
    });

    it("handles query strings", () => {
      expect(cleanFileName("http://localhost:5173/src/App.tsx?t=123456789")).toBe("App.tsx");
    });

    it("handles hashes", () => {
      expect(cleanFileName("http://localhost:5173/src/App.tsx#section")).toBe("App.tsx");
    });

    it("handles system paths", () => {
      expect(cleanFileName("C:\\projects\\ostenia\\src\\main.tsx")).toBe("main.tsx");
      expect(cleanFileName("/home/user/ostenia/src/main.tsx")).toBe("main.tsx");
    });

    it("handles empty or falsy values", () => {
      expect(cleanFileName("")).toBe("");
    });
  });

  describe("parseStackTrace", () => {
    it("returns empty string and undefined caller for empty stack", () => {
      const result = parseStackTrace("");
      expect(result.rawStack).toBe("");
      expect(result.caller).toBeUndefined();
    });

    it("parses V8 stack trace format with function name", () => {
      const stack = `Error
    at addLog (http://localhost:5173/src/App.tsx:130:15)
    at handleToggleService (http://localhost:5173/src/App.tsx:392:5)
    at HTMLButtonElement.onClick (http://localhost:5173/src/components/ServiceItem.tsx:20:10)`;

      const result = parseStackTrace(stack);
      expect(result.caller).toBeDefined();
      expect(result.caller?.functionName).toBe("handleToggleService");
      expect(result.caller?.fileName).toBe("App.tsx");
      expect(result.caller?.line).toBe("392");
      expect(result.caller?.column).toBe("5");
    });

    it("parses V8 stack trace format without function name (anonymous)", () => {
      const stack = `Error
    at http://localhost:5173/src/App.tsx:392:5`;

      const result = parseStackTrace(stack);
      expect(result.caller).toBeDefined();
      expect(result.caller?.functionName).toBe("anonymous");
      expect(result.caller?.fileName).toBe("App.tsx");
      expect(result.caller?.line).toBe("392");
      expect(result.caller?.column).toBe("5");
    });

    it("parses Firefox/Safari stack trace format", () => {
      const stack = `addLog@http://localhost:5173/src/App.tsx:130:15
handleToggleService@http://localhost:5173/src/App.tsx:392:5
onClick@http://localhost:5173/src/components/ServiceItem.tsx:20:10`;

      const result = parseStackTrace(stack);
      expect(result.caller).toBeDefined();
      expect(result.caller?.functionName).toBe("handleToggleService");
      expect(result.caller?.fileName).toBe("App.tsx");
      expect(result.caller?.line).toBe("392");
      expect(result.caller?.column).toBe("5");
    });

    it("handles ignore list filtering to skip logging wrapper frames", () => {
      const stack = `Error
    at parseStackTrace (http://localhost:5173/src/utils/stackParser.ts:35:10)
    at addLog (http://localhost:5173/src/App.tsx:130:15)
    at Object.console.log (http://localhost:5173/src/App.tsx:245:5)
    at myCoolAppFunc (http://localhost:5173/src/components/SettingsModal.tsx:40:12)`;

      const result = parseStackTrace(stack);
      expect(result.caller).toBeDefined();
      expect(result.caller?.functionName).toBe("myCoolAppFunc");
      expect(result.caller?.fileName).toBe("SettingsModal.tsx");
    });
  });
});
