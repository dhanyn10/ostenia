import { describe, it, expect } from "vitest";
import { shouldIgnoreCardToggle } from "@/utils/utils";

describe("dom utils", () => {
  describe("shouldIgnoreCardToggle", () => {
    it("returns true when target is a standard button without data-card-expand", () => {
      const button = document.createElement("button");
      expect(shouldIgnoreCardToggle(button)).toBe(true);
    });

    it("returns true when target is inside a standard button without data-card-expand", () => {
      const button = document.createElement("button");
      const icon = document.createElement("svg");
      button.appendChild(icon);
      expect(shouldIgnoreCardToggle(icon)).toBe(true);
    });

    it("returns false when target is a button with data-card-expand", () => {
      const button = document.createElement("button");
      button.setAttribute("data-card-expand", "true");
      expect(shouldIgnoreCardToggle(button)).toBe(false);
    });

    it("returns false when target is inside a button with data-card-expand", () => {
      const button = document.createElement("button");
      button.setAttribute("data-card-expand", "true");
      const span = document.createElement("span");
      button.appendChild(span);
      expect(shouldIgnoreCardToggle(span)).toBe(false);
    });

    it("returns true when target is an anchor tag or inside an anchor tag", () => {
      const link = document.createElement("a");
      const span = document.createElement("span");
      link.appendChild(span);

      expect(shouldIgnoreCardToggle(link)).toBe(true);
      expect(shouldIgnoreCardToggle(span)).toBe(true);
    });

    it("returns true when target is an input element", () => {
      const input = document.createElement("input");
      expect(shouldIgnoreCardToggle(input)).toBe(true);
    });

    it("returns true when target is or is inside an element with data-no-card-toggle", () => {
      const container = document.createElement("div");
      container.setAttribute("data-no-card-toggle", "true");
      const child = document.createElement("div");
      const innerSpan = document.createElement("span");
      child.appendChild(innerSpan);
      container.appendChild(child);

      expect(shouldIgnoreCardToggle(container)).toBe(true);
      expect(shouldIgnoreCardToggle(child)).toBe(true);
      expect(shouldIgnoreCardToggle(innerSpan)).toBe(true);
    });

    it("returns false for regular elements (div, span) not inside interactive elements", () => {
      const container = document.createElement("div");
      const span = document.createElement("span");
      container.appendChild(span);

      expect(shouldIgnoreCardToggle(container)).toBe(false);
      expect(shouldIgnoreCardToggle(span)).toBe(false);
    });
  });
});
