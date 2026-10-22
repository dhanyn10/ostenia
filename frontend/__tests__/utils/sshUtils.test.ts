import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  isArchiveFile,
  calculateContextMenuY,
  isIgnorableSFTPError,
  getSSHConfigFromStorage,
  extractPathFromTitle,
  formatSize,
  filterAndSortFiles,
  setupZoomKeyHandler,
} from "@/utils/utils";

describe("sshUtils", () => {
  describe("isArchiveFile", () => {
    it("returns true for known archive extensions", () => {
      expect(isArchiveFile("archive.zip")).toBe(true);
      expect(isArchiveFile("backup.tar.gz")).toBe(true);
      expect(isArchiveFile("files.7z")).toBe(true);
      expect(isArchiveFile("data.rar")).toBe(true);
      expect(isArchiveFile("data.tar")).toBe(true);
      expect(isArchiveFile("data.bz2")).toBe(true);
      expect(isArchiveFile("data.xz")).toBe(true);
    });

    it("returns false for non-archive extensions or empty/undefined", () => {
      expect(isArchiveFile("document.pdf")).toBe(false);
      expect(isArchiveFile("script.sh")).toBe(false);
      expect(isArchiveFile("")).toBe(false);
      expect(isArchiveFile(undefined)).toBe(false);
    });
  });

  describe("calculateContextMenuY", () => {
    it("returns clientY if within window height", () => {
      window.innerHeight = 1000;
      expect(calculateContextMenuY(200, 150)).toBe(200);
    });

    it("adjusts Y to avoid overflowing window bottom", () => {
      window.innerHeight = 800;
      // 700 + 150 = 850 > 800
      expect(calculateContextMenuY(700, 150)).toBe(550);
    });

    it("clamps minimum Y to 10", () => {
      window.innerHeight = 100;
      // 90 + 150 = 240 > 100 -> 90 - 150 = -60 -> clamped to 10
      expect(calculateContextMenuY(90, 150)).toBe(10);
    });
  });

  describe("isIgnorableSFTPError", () => {
    it("returns true for common transient / session closed sftp errors", () => {
      expect(isIgnorableSFTPError("EOF received")).toBe(true);
      expect(isIgnorableSFTPError(new Error("session not found"))).toBe(true);
      expect(isIgnorableSFTPError("Session not connected")).toBe(true);
      expect(isIgnorableSFTPError("sftp not connected")).toBe(true);
    });

    it("returns false for actual application or network errors", () => {
      expect(isIgnorableSFTPError("Permission denied")).toBe(false);
      expect(isIgnorableSFTPError("File not found")).toBe(false);
    });
  });

  describe("getSSHConfigFromStorage", () => {
    beforeEach(() => {
      localStorage.clear();
    });

    it("returns default values when localStorage is empty", () => {
      const config = getSSHConfigFromStorage();
      expect(config.maxTimeout).toBe(10);
      expect(config.maxRetries).toBe(3);
    });

    it("reads valid stored values", () => {
      localStorage.setItem("ostenia_ssh_max_timeout", "25");
      localStorage.setItem("ostenia_ssh_max_retries", "5");
      const config = getSSHConfigFromStorage();
      expect(config.maxTimeout).toBe(25);
      expect(config.maxRetries).toBe(5);
    });

    it("falls back to defaults if stored values are invalid or zero/negative", () => {
      localStorage.setItem("ostenia_ssh_max_timeout", "-5");
      localStorage.setItem("ostenia_ssh_max_retries", "abc");
      const config = getSSHConfigFromStorage();
      expect(config.maxTimeout).toBe(10);
      expect(config.maxRetries).toBe(3);
    });
  });

  describe("extractPathFromTitle", () => {
    it("extracts path after colon if it starts with slash", () => {
      expect(extractPathFromTitle("user@host: /var/www/html")).toBe(
        "/var/www/html",
      );
      expect(extractPathFromTitle("Terminal: /home/ubuntu")).toBe(
        "/home/ubuntu",
      );
    });

    it("returns null if no colon or path does not start with slash", () => {
      expect(extractPathFromTitle("Bash terminal")).toBeNull();
      expect(extractPathFromTitle("user@host: ~")).toBeNull();
      expect(extractPathFromTitle("user@host: relative/path")).toBeNull();
    });
  });

  describe("formatSize", () => {
    it("returns - for 0 or falsy bytes", () => {
      expect(formatSize(0)).toBe("-");
    });

    it("formats bytes, KB, MB, GB properly", () => {
      expect(formatSize(500)).toBe("500.0 B");
      expect(formatSize(1024)).toBe("1.0 KB");
      expect(formatSize(1024 * 1024 * 2.5)).toBe("2.5 MB");
      expect(formatSize(1024 * 1024 * 1024 * 10)).toBe("10.0 GB");
    });
  });

  describe("filterAndSortFiles", () => {
    const mockFiles = [
      { name: "b_file.txt", isDir: false, size: 200 },
      { name: "a_folder", isDir: true, size: 0 },
      { name: ".hidden", isDir: false, size: 50 },
      { name: "c_file.zip", isDir: false, size: 1000 },
    ];

    it("puts directories first, then sorts by name ascending", () => {
      const sorted = filterAndSortFiles(mockFiles, "", true, {
        key: "name",
        direction: "asc",
      });
      expect(sorted[0].name).toBe("a_folder");
      expect(sorted[1].name).toBe(".hidden");
      expect(sorted[2].name).toBe("b_file.txt");
      expect(sorted[3].name).toBe("c_file.zip");
    });

    it("filters out hidden files when showHiddenFiles is false", () => {
      const filtered = filterAndSortFiles(mockFiles, "", false, {
        key: "name",
        direction: "asc",
      });
      expect(filtered.some((f) => f.name === ".hidden")).toBe(false);
      expect(filtered).toHaveLength(3);
    });

    it("filters files by search query (case-insensitive)", () => {
      const filtered = filterAndSortFiles(mockFiles, "FILE", true, {
        key: "name",
        direction: "asc",
      });
      expect(filtered).toHaveLength(2);
      expect(filtered.map((f) => f.name)).toEqual(["b_file.txt", "c_file.zip"]);
    });

    it("sorts files by size descending", () => {
      const sorted = filterAndSortFiles(mockFiles, "", true, {
        key: "size",
        direction: "desc",
      });
      expect(sorted[0].name).toBe("a_folder"); // dir always first
      expect(sorted[1].name).toBe("c_file.zip"); // 1000
      expect(sorted[2].name).toBe("b_file.txt"); // 200
      expect(sorted[3].name).toBe(".hidden"); // 50
    });
  });

  describe("setupZoomKeyHandler", () => {
    it("handles zoom keys when enabled", () => {
      localStorage.setItem("ostenia_ssh_zoom_enabled", "true");
      let attachedHandler: ((event: KeyboardEvent) => boolean) | null = null;
      const mockTerm = {
        attachCustomKeyEventHandler: vi.fn((fn) => {
          attachedHandler = fn;
        }),
      } as any;

      const handlers = {
        handleZoomIn: vi.fn(),
        handleZoomOut: vi.fn(),
        handleZoomReset: vi.fn(),
      };
      const handlersRef = { current: handlers } as any;

      setupZoomKeyHandler(mockTerm, handlersRef);
      expect(mockTerm.attachCustomKeyEventHandler).toHaveBeenCalled();

      // Test Zoom In (Ctrl + =)
      const zoomInEvent = new KeyboardEvent("keydown", {
        key: "=",
        ctrlKey: true,
      });
      const resultIn = attachedHandler!(zoomInEvent);
      expect(resultIn).toBe(false);
      expect(handlers.handleZoomIn).toHaveBeenCalledTimes(1);

      // Test Zoom Out (Ctrl + -)
      const zoomOutEvent = new KeyboardEvent("keydown", {
        key: "-",
        ctrlKey: true,
      });
      const resultOut = attachedHandler!(zoomOutEvent);
      expect(resultOut).toBe(false);
      expect(handlers.handleZoomOut).toHaveBeenCalledTimes(1);

      // Test Zoom Reset (Ctrl + 0)
      const zoomResetEvent = new KeyboardEvent("keydown", {
        key: "0",
        ctrlKey: true,
      });
      const resultReset = attachedHandler!(zoomResetEvent);
      expect(resultReset).toBe(false);
      expect(handlers.handleZoomReset).toHaveBeenCalledTimes(1);

      // Test other key
      const otherEvent = new KeyboardEvent("keydown", {
        key: "a",
        ctrlKey: true,
      });
      expect(attachedHandler!(otherEvent)).toBe(true);
    });

    it("ignores zoom keys when disabled in localStorage", () => {
      localStorage.setItem("ostenia_ssh_zoom_enabled", "false");
      let attachedHandler: ((event: KeyboardEvent) => boolean) | null = null;
      const mockTerm = {
        attachCustomKeyEventHandler: vi.fn((fn) => {
          attachedHandler = fn;
        }),
      } as any;

      const handlers = {
        handleZoomIn: vi.fn(),
        handleZoomOut: vi.fn(),
        handleZoomReset: vi.fn(),
      };
      const handlersRef = { current: handlers } as any;

      setupZoomKeyHandler(mockTerm, handlersRef);

      const zoomInEvent = new KeyboardEvent("keydown", {
        key: "+",
        ctrlKey: true,
      });
      const result = attachedHandler!(zoomInEvent);
      expect(result).toBe(true);
      expect(handlers.handleZoomIn).not.toHaveBeenCalled();
    });
  });
});
