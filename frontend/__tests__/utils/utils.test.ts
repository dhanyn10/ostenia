import { describe, it, expect } from "vitest";
import { cn } from "@/lib/utils";

describe("cn utility function", () => {
  it("merges class names correctly", () => {
    expect(cn("px-2", "py-1")).toBe("px-2 py-1");
  });

  it("handles conditional class names", () => {
    const isHidden = false;
    const isPrimary = true;
    expect(cn("btn", isHidden && "hidden", isPrimary && "btn-primary")).toBe(
      "btn btn-primary",
    );
  });

  it("resolves tailwind conflicts using tailwind-merge", () => {
    expect(cn("px-2 text-red-500", "px-4 text-blue-500")).toBe(
      "px-4 text-blue-500",
    );
  });

  it("handles arrays and undefined/null inputs", () => {
    expect(cn(["px-2", null, undefined, false, "py-1"])).toBe("px-2 py-1");
  });
});
