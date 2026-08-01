import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { getCallerInfo } from "./stackParser";

describe("stackParser", () => {
  const originalError = globalThis.Error;

  afterEach(() => {
    globalThis.Error = originalError;
  });

  it("parses Chrome-style stack traces correctly", () => {
    const fakeStack = `Error
    at addLog (http://localhost:5173/src/App.tsx:118:10)
    at console.error (http://localhost:5173/src/App.tsx:273:14)
    at handleToggleService (http://localhost:5173/src/App.tsx:392:5)
    at HTMLButtonElement.onClick (http://localhost:5173/src/components/ServiceItem.tsx:10:20)`;

    // Mock Error constructor
    class MockError {
      stack = fakeStack;
    }
    globalThis.Error = MockError as any;

    const info = getCallerInfo();
    expect(info.functionName).toBe("handleToggleService");
    expect(info.fileName).toBe("App.tsx");
    expect(info.line).toBe("392");
    expect(info.column).toBe("5");
  });

  it("parses Firefox-style stack traces correctly", () => {
    const fakeStack = `
getCallerInfo@http://localhost:5173/src/utils/stackParser.ts:10:5
addLog@http://localhost:5173/src/App.tsx:120:10
fetchServiceStatus@http://localhost:5173/src/App.tsx:42:15
`;

    class MockError {
      stack = fakeStack;
    }
    globalThis.Error = MockError as any;

    const info = getCallerInfo();
    expect(info.functionName).toBe("fetchServiceStatus");
    expect(info.fileName).toBe("App.tsx");
    expect(info.line).toBe("42");
    expect(info.column).toBe("15");
  });

  it("handles query parameters in URLs", () => {
    const fakeStack = `Error
    at addLog (http://localhost:5173/src/App.tsx?t=172348329:118:10)
    at console.error (http://localhost:5173/src/App.tsx?t=172348329:273:14)
    at loadInitialData (http://localhost:5173/src/App.tsx?t=172348329:180:22)`;

    class MockError {
      stack = fakeStack;
    }
    globalThis.Error = MockError as any;

    const info = getCallerInfo();
    expect(info.functionName).toBe("loadInitialData");
    expect(info.fileName).toBe("App.tsx");
    expect(info.line).toBe("180");
    expect(info.column).toBe("22");
  });

  it("returns fallback values when stack is empty", () => {
    class MockError {
      stack = "";
    }
    globalThis.Error = MockError as any;

    const info = getCallerInfo();
    expect(info.functionName).toBe("unknown");
    expect(info.fileName).toBe("unknown");
    expect(info.line).toBe("0");
    expect(info.column).toBe("0");
  });
});
