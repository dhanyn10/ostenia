import { describe, it, expect } from "vitest";
import { getServiceActionFlags } from "@/utils/utils";

describe("serviceHelpers", () => {
  describe("getServiceActionFlags", () => {
    it("determines installation status correctly", () => {
      const service = { name: "MySQL" };
      const taskWithVersions = { installedVers: ["8.0.30"] };
      const taskWithoutVersions = { installedVers: [] };

      const flagsInstalled = getServiceActionFlags(
        service,
        taskWithVersions,
        false,
        false,
        false,
      );
      expect(flagsInstalled.isInstalled).toBe(true);

      const flagsNotInstalled = getServiceActionFlags(
        service,
        taskWithoutVersions,
        false,
        false,
        false,
      );
      expect(flagsNotInstalled.isInstalled).toBe(false);
    });

    it("treats OpenSSL as installed regardless of task.installedVers", () => {
      const service = { name: "OpenSSL" };
      const task = { installedVers: [] };

      const flags = getServiceActionFlags(service, task, false, false, false);
      expect(flags.isInstalled).toBe(true);
      expect(flags.hasTerminal).toBe(false);
      expect(flags.hasOpenFolder).toBe(false);
    });

    it("enables HTTPS toggle for Apache or Nginx when OpenSSL is enabled", () => {
      const apacheService = { name: "Apache" };
      const nginxService = { name: "Nginx" };
      const task = { installedVers: ["2.4"] };

      const apacheFlags = getServiceActionFlags(
        apacheService,
        task,
        true,
        false,
        true,
      );
      expect(apacheFlags.hasHttpsToggle).toBe(true);
      expect(apacheFlags.isHttpsEnabled).toBe(true);

      const nginxFlags = getServiceActionFlags(
        nginxService,
        task,
        false,
        true,
        true,
      );
      expect(nginxFlags.hasHttpsToggle).toBe(true);
      expect(nginxFlags.isHttpsEnabled).toBe(true);

      const noSslFlags = getServiceActionFlags(
        apacheService,
        task,
        true,
        false,
        false,
      );
      expect(noSslFlags.hasHttpsToggle).toBe(false);
    });

    it("configures special actions for PHP and HeidiSQL", () => {
      const php = { name: "PHP" };
      const phpFlags = getServiceActionFlags(
        php,
        { installedVers: ["8.2"] },
        false,
        false,
        false,
      );
      expect(phpFlags.hasPhpExtManager).toBe(true);
      expect(phpFlags.hasTerminal).toBe(true);
      expect(phpFlags.hasOpenFolder).toBe(true);

      const heidi = { name: "HeidiSQL" };
      const heidiFlags = getServiceActionFlags(
        heidi,
        { installedVers: ["12.0"] },
        false,
        false,
        false,
      );
      expect(heidiFlags.hasHeidiOpen).toBe(true);
      expect(heidiFlags.hasTerminal).toBe(false);
      expect(heidiFlags.hasOpenFolder).toBe(false);
    });
  });
});
