import { describe, it, expect } from "vitest";
import {
  resolvePrerequisiteTask,
  isCustomVersionAllowed,
  isParentTaskActive,
  isTaskVersionInstalled,
} from "@/utils/utils";

describe("pluginHelpers", () => {
  describe("resolvePrerequisiteTask", () => {
    it("correctly formats task target and version for PHP", () => {
      const task = {
        name: "PHP",
        version: "8.1",
        versionUrls: { "8.2": "http://example.com/php82.zip" },
      };
      const resolved = resolvePrerequisiteTask(task, "8.2");

      expect(resolved.version).toBe("8.2");
      expect(resolved.target).toBe("php/php-8.2");
      expect(resolved.url).toBe("http://example.com/php82.zip");
    });

    it("correctly maps special category for Node.js and HeidiSQL", () => {
      const nodeTask = { name: "Node.js", version: "18.0.0" };
      expect(resolvePrerequisiteTask(nodeTask).target).toBe(
        "nodejs/node-v18.0.0",
      );

      const heidiTask = { name: "HeidiSQL", version: "12.0" };
      expect(resolvePrerequisiteTask(heidiTask).target).toBe("heidisql/12.0");
    });

    it("falls back to default task version if selectedVersion is omitted", () => {
      const apacheTask = { name: "Apache", version: "2.4" };
      const resolved = resolvePrerequisiteTask(apacheTask);
      expect(resolved.version).toBe("2.4");
      expect(resolved.target).toBe("apache/httpd-2.4");
    });
  });

  describe("isCustomVersionAllowed", () => {
    it("returns false for HeidiSQL and OpenSSL", () => {
      expect(isCustomVersionAllowed("HeidiSQL")).toBe(false);
      expect(isCustomVersionAllowed("OpenSSL")).toBe(false);
    });

    it("returns true for other tasks like PHP, Apache, Node.js", () => {
      expect(isCustomVersionAllowed("PHP")).toBe(true);
      expect(isCustomVersionAllowed("Apache")).toBe(true);
      expect(isCustomVersionAllowed("Node.js")).toBe(true);
    });
  });

  describe("isParentTaskActive", () => {
    it("returns true when task is in progress", () => {
      expect(isParentTaskActive({ status: "Downloading" })).toBe(true);
      expect(isParentTaskActive({ status: "Extracting" })).toBe(true);
    });

    it("returns false when task is Ready, Completed, or starts with Error", () => {
      expect(isParentTaskActive({ status: "Ready" })).toBe(false);
      expect(isParentTaskActive({ status: "Completed" })).toBe(false);
      expect(isParentTaskActive({ status: "Error: Failed to fetch" })).toBe(
        false,
      );
      expect(isParentTaskActive(null)).toBe(false);
      expect(isParentTaskActive(undefined)).toBe(false);
    });
  });

  describe("isTaskVersionInstalled", () => {
    it("handles OpenSSL and HeidiSQL based on isInstalled flag", () => {
      expect(
        isTaskVersionInstalled({ name: "OpenSSL", isInstalled: true }),
      ).toBe(true);
      expect(
        isTaskVersionInstalled({ name: "OpenSSL", isInstalled: false }),
      ).toBe(false);
      expect(
        isTaskVersionInstalled({ name: "HeidiSQL", isInstalled: true }),
      ).toBe(true);
    });

    it("checks installedVersions array for other plugins", () => {
      const task = { name: "PHP", version: "8.1" };
      const installed = ["8.0", "8.1", "8.2"];

      expect(isTaskVersionInstalled(task, "8.1", installed)).toBe(true);
      expect(isTaskVersionInstalled(task, "8.3", installed)).toBe(false);
      expect(isTaskVersionInstalled(task, undefined, installed)).toBe(true);
    });
  });
});
