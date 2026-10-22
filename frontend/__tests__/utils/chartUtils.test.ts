import { describe, it, expect } from "vitest";
import {
  buildSparklineSegments,
  buildSparklinePaths,
  type ResourceDataPoint,
} from "@/utils/chartUtils";

describe("chartUtils", () => {
  describe("buildSparklineSegments", () => {
    it("handles non-null sequential data points", () => {
      const data: ResourceDataPoint[] = [
        { cpu: 0, mem: 10, disk: 20 },
        { cpu: 50, mem: 50, disk: 50 },
        { cpu: 100, mem: 90, disk: 80 },
      ];

      const segments = buildSparklineSegments(data, "cpu", 100, 50, 3);
      expect(segments).toHaveLength(1);
      expect(segments[0]).toHaveLength(3);

      // Check X positions: 0, 50, 100
      expect(segments[0][0][0]).toBe(0);
      expect(segments[0][1][0]).toBe(50);
      expect(segments[0][2][0]).toBe(100);
    });

    it("splits segments when null values appear", () => {
      const data: ResourceDataPoint[] = [
        { cpu: 20, mem: null, disk: null },
        { cpu: null, mem: null, disk: null },
        { cpu: 60, mem: null, disk: null },
      ];

      const segments = buildSparklineSegments(data, "cpu", 100, 50, 3);
      expect(segments).toHaveLength(2);
      expect(segments[0]).toHaveLength(1);
      expect(segments[1]).toHaveLength(1);
    });

    it("returns empty array when all data points are null", () => {
      const data: ResourceDataPoint[] = [
        { cpu: null, mem: null, disk: null },
        { cpu: null, mem: null, disk: null },
      ];

      const segments = buildSparklineSegments(data, "cpu", 100, 50, 2);
      expect(segments).toHaveLength(0);
    });
  });

  describe("buildSparklinePaths", () => {
    it("generates correct M, L, and Z path instructions", () => {
      const segments: Array<Array<[number, number]>> = [
        [
          [0, 40],
          [50, 20],
          [100, 10],
        ],
      ];

      const { linePath, areaPath } = buildSparklinePaths(segments, 50);

      expect(linePath.trim()).toBe("M 0 40 L 50 20 L 100 10");
      expect(areaPath.trim()).toBe("M 0 50 L 0 40 L 50 20 L 100 10 L 100 50 Z");
    });

    it("returns empty strings for empty segments", () => {
      const { linePath, areaPath } = buildSparklinePaths([], 50);
      expect(linePath).toBe("");
      expect(areaPath).toBe("");
    });

    it("skips empty inner segments", () => {
      const { linePath, areaPath } = buildSparklinePaths([[]], 50);
      expect(linePath).toBe("");
      expect(areaPath).toBe("");
    });
  });
});
