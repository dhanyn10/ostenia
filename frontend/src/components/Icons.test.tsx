import React from "react";
import { render } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import Icons from "./Icons";

// Mock the static assets to avoid any path resolution issue
vi.mock("../assets/icons/plugins.svg", () => ({
  default: "mock-plugins.svg",
}));

describe("Icons Component", () => {
  describe("RawSVGIcon (Icons.Raw)", () => {
    it("returns null if svgString is empty", () => {
      const { container } = render(<Icons.Raw svgString="" />);
      expect(container.firstChild).toBeNull();
    });

    it("renders with the correct style masks and size when svgString is provided", () => {
      const svgString = '<svg><path d="M10 10h10v10H10z"/></svg>';
      const { container } = render(
        <Icons.Raw svgString={svgString} size={24} className="test-class" />,
      );

      const div = container.firstChild as HTMLElement;
      expect(div).toBeInTheDocument();
      expect(div.className).toContain("test-class");
      expect(div.style.width).toBe("24px");
      expect(div.style.height).toBe("24px");

      // Check the mask styles
      const encodedSvg = encodeURIComponent(svgString);
      expect(div.style.maskImage).toContain(`data:image/svg+xml,${encodedSvg}`);
      expect(div.style.webkitMaskImage).toContain(
        `data:image/svg+xml,${encodedSvg}`,
      );
      expect(div.style.maskRepeat).toBe("no-repeat");
      expect(div.style.maskPosition).toBe("center");
      expect(div.style.maskSize).toBe("contain");
    });

    it("cleans <?xml ?> tag from svgString", () => {
      const svgString =
        '<?xml version="1.0" encoding="UTF-8"?><svg><rect /></svg>';
      const { container } = render(<Icons.Raw svgString={svgString} />);

      const div = container.firstChild as HTMLElement;
      expect(div).toBeInTheDocument();

      const expectedCleanSvg = "<svg><rect /></svg>";
      const encodedSvg = encodeURIComponent(expectedCleanSvg);
      expect(div.style.maskImage).toContain(encodedSvg);
      expect(div.style.maskImage).not.toContain("?xml");
    });
  });

  describe("Plugins (Icons.Plugins)", () => {
    it("renders correctly with src and default size", () => {
      const { container } = render(<Icons.Plugins className="plugin-icon" />);

      const div = container.firstChild as HTMLElement;
      expect(div).toBeInTheDocument();
      expect(div.className).toContain("plugin-icon");
      expect(div.style.width).toBe("20px");
      expect(div.style.height).toBe("20px");
      expect(div.style.maskImage).toContain("mock-plugins.svg");
    });
  });
});
