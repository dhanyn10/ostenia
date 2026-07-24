import {
  render,
  screen,
  fireEvent,
  waitFor,
  act,
} from "@testing-library/react";
import { vi, describe, it, expect, beforeEach } from "vitest";
import SSHSessionView from "./SSHSessionView";
import * as AppBackend from "../../wailsjs/go/backend/App";

// Mock xterm
vi.mock("@xterm/xterm", () => {
  return {
    Terminal: vi.fn().mockImplementation(() => ({
      loadAddon: vi.fn(),
      open: vi.fn(),
      dispose: vi.fn(),
      onData: vi.fn(),
      onTitleChange: vi.fn(),
      write: vi.fn(),
      options: { theme: {} },
      focus: vi.fn(),
      clear: vi.fn(),
      getSelection: vi.fn().mockReturnValue("selected terminal text"),
    })),
  };
});

vi.mock("@xterm/addon-fit", () => {
  return {
    FitAddon: vi.fn().mockImplementation(() => ({
      fit: vi.fn(),
      proposeDimensions: vi.fn().mockReturnValue({ cols: 120, rows: 24 }),
    })),
  };
});

const eventCallbacks: Record<string, Function> = {};
vi.mock("../../wailsjs/runtime/runtime", () => ({
  EventsOn: vi.fn((event, cb) => {
    eventCallbacks[event] = cb;
  }),
  EventsOff: vi.fn(),
}));

// Mock Wails backend functions
vi.mock("../../wailsjs/go/backend/App", () => ({
  ConnectSSH: vi.fn().mockResolvedValue(null),
  GetRemoteFiles: vi.fn().mockResolvedValue([
    { name: "test.txt", isDir: false, size: 1024 },
    { name: "folder", isDir: true, size: 0 },
    { name: ".env", isDir: false, size: 256 },
    { name: ".hidden_dir", isDir: true, size: 0 },
    { name: "archive.zip", isDir: false, size: 50000 },
  ]),
  GetRemoteCurrentPath: vi.fn().mockResolvedValue("/home/user"),
  SendSSHInput: vi.fn().mockResolvedValue(null),
  ResizeSSHTerminal: vi.fn().mockResolvedValue(null),
  DownloadRemoteFile: vi.fn().mockResolvedValue(null),
  UploadRemoteFile: vi.fn().mockResolvedValue(null),
  EditRemoteFile: vi.fn().mockResolvedValue(null),
  ExecuteSFTPAction: vi.fn().mockResolvedValue(null),
  GetSSHResourceUsage: vi.fn().mockResolvedValue({
    cpu: 0,
    mem: 0,
    memTotal: 8192,
    memUsed: 0,
    disk: 0,
    diskTotal: 102400,
    diskUsed: 0,
  }),
}));

describe("SSHSessionView Component", () => {
  const mockSession = {
    id: "session-1",
    name: "Test Server",
    host: "1.2.3.4",
    user: "root",
  };

  const mockProps = {
    session: mockSession,
    onClose: vi.fn(),
    addToast: vi.fn(),
    isActive: true,
    theme: "light",
    onOpenSettings: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    window.confirm = vi.fn().mockReturnValue(true);
    window.prompt = vi.fn().mockReturnValue("new-name");

    // Mock navigator.clipboard
    Object.defineProperty(navigator, "clipboard", {
      writable: true,
      configurable: true,
      value: {
        writeText: vi.fn().mockResolvedValue(undefined),
        readText: vi.fn().mockResolvedValue("pasted text"),
      },
    });
  });

  it("renders and connects to SSH", async () => {
    render(<SSHSessionView {...mockProps} />);

    expect(AppBackend.ConnectSSH).toHaveBeenCalledWith(mockSession);

    await waitFor(() => {
      expect(screen.queryByText(/Connecting\.\.\./i)).not.toBeInTheDocument();
    });
  });

  it("lists remote files after connection", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(AppBackend.GetRemoteFiles).toHaveBeenCalled();
    });

    expect(screen.getByText("test.txt")).toBeInTheDocument();
    expect(screen.getByText("folder")).toBeInTheDocument();
  });

  it("handles toolbar actions", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.queryByText(/Connecting\.\.\./i)).not.toBeInTheDocument();
    });

    const toggleExplorerBtn = screen.getByTitle(/Toggle Explorer/i);
    fireEvent.click(toggleExplorerBtn);
    expect(screen.queryByText("test.txt")).not.toBeInTheDocument();
  });

  it("handles file deletion", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    const fileItem = screen.getByText("test.txt");
    fireEvent.contextMenu(fileItem);

    const deleteBtn = screen.getByText(/Delete/i);
    fireEvent.click(deleteBtn);

    expect(window.confirm).toHaveBeenCalled();
    expect(AppBackend.ExecuteSFTPAction).toHaveBeenCalledWith(
      mockSession.id,
      "delete",
      expect.stringContaining("test.txt"),
      "",
    );
  });

  it("handles directory navigation", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("folder")).toBeInTheDocument();
    });

    const folderItem = screen.getByText("folder");
    fireEvent.doubleClick(folderItem);

    expect(AppBackend.GetRemoteFiles).toHaveBeenCalledWith(
      mockSession.id,
      expect.stringContaining("folder"),
    );
  });

  it("shows hidden files by default and filters them after toggling", async () => {
    render(<SSHSessionView {...mockProps} />);

    // Wait for file listing to load
    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    // Hidden files should be visible by default
    expect(screen.getByText(".env")).toBeInTheDocument();

    // Right click on the file explorer background to open the context menu
    const nameHeader = screen.getByText("Name");
    fireEvent.contextMenu(nameHeader);

    // Context menu should appear
    const toggleHiddenBtn = screen.getByText("View hidden files/folder");
    expect(toggleHiddenBtn).toBeInTheDocument();

    // Click "View hidden files/folder"
    fireEvent.click(toggleHiddenBtn);

    // Hidden files should now be hidden!
    await waitFor(() => {
      expect(screen.queryByText(".env")).not.toBeInTheDocument();
    });
  });

  it("adjusts context menu position to prevent bottom cut-offs", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    // Mock window innerHeight
    const originalInnerHeight = window.innerHeight;
    Object.defineProperty(window, "innerHeight", {
      writable: true,
      configurable: true,
      value: 600,
    });

    // 1. Test explorer context menu near the bottom
    const nameHeader = screen.getByText("Name");
    fireEvent.contextMenu(nameHeader, { clientY: 550, clientX: 100 });

    // Menu should be rendered. Its Y position should be offset upwards
    const menuElement = screen.getByText("View hidden files/folder").closest("div");
    expect(menuElement).toBeInTheDocument();
    // Expected Y position: 550 - 130 = 420
    expect(menuElement?.style.top).toBe("420px");

    // Close menu by clicking
    fireEvent.click(document.body);

    // 2. Test file context menu near the bottom
    const fileItem = screen.getByText("test.txt");
    fireEvent.contextMenu(fileItem, { clientY: 500, clientX: 100 });

    const fileMenuElement = screen.getByText("Rename").closest("div");
    expect(fileMenuElement).toBeInTheDocument();
    // Expected Y position: 500 - 180 = 320
    expect(fileMenuElement?.style.top).toBe("320px");

    // Restore innerHeight
    Object.defineProperty(window, "innerHeight", {
      writable: true,
      configurable: true,
      value: originalInnerHeight,
    });
  });

  it("handles explorer context menu 'Back' and 'Refresh' actions", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    // Open background context menu
    const nameHeader = screen.getByText("Name");
    fireEvent.contextMenu(nameHeader);

    // Test Refresh action
    const refreshBtn = screen.getByText("Refresh");
    fireEvent.click(refreshBtn);

    // Expect another GetRemoteFiles fetch call
    expect(AppBackend.GetRemoteFiles).toHaveBeenCalled();

    // Re-open background context menu to test Back action
    fireEvent.contextMenu(nameHeader);
    const backBtn = screen.getByText("Back");
    expect(backBtn).toBeInTheDocument();

    // Since default mock current path is not empty/root in nested calls, trigger Back
    fireEvent.click(backBtn);
  });

  it("handles file download, edit, and open settings", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    const fileItem = screen.getByText("test.txt");
    fireEvent.contextMenu(fileItem);

    // 1. Download
    const downloadBtn = screen.getByText("Download");
    fireEvent.click(downloadBtn);
    await waitFor(() => {
      expect(AppBackend.DownloadRemoteFile).toHaveBeenCalledWith(mockSession.id, "/home/user/test.txt");
    });

    // 2. Open With
    fireEvent.contextMenu(fileItem);
    const openWithBtn = screen.getByText("Open With");
    fireEvent.click(openWithBtn);
    expect(mockProps.onOpenSettings).toHaveBeenCalledWith("config");

    // 3. Edit (via context menu)
    fireEvent.contextMenu(fileItem);
    const editBtn = screen.getByText("Edit File");
    fireEvent.click(editBtn);
    await waitFor(() => {
      expect(AppBackend.EditRemoteFile).toHaveBeenCalledWith(mockSession.id, "/home/user/test.txt");
      expect(mockProps.addToast).toHaveBeenCalledWith("Success", "File saved and uploaded", "success");
    });

    // 4. Edit (via double click on file)
    fireEvent.doubleClick(fileItem);
    await waitFor(() => {
      expect(AppBackend.EditRemoteFile).toHaveBeenCalled();
    });
  });

  it("handles Upload and New Folder buttons", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    // 1. Upload
    const uploadBtn = screen.getByText("Upload");
    fireEvent.click(uploadBtn);
    await waitFor(() => {
      expect(AppBackend.UploadRemoteFile).toHaveBeenCalledWith(mockSession.id, "/home/user");
      expect(mockProps.addToast).toHaveBeenCalledWith("Success", "File uploaded successfully", "success");
    });

    // 2. New Folder
    window.prompt = vi.fn().mockReturnValue("new_folder_name");
    const newBtn = screen.getByText("New");
    fireEvent.click(newBtn);
    await waitFor(() => {
      expect(AppBackend.ExecuteSFTPAction).toHaveBeenCalledWith(mockSession.id, "mkdir", "/home/user/new_folder_name", "");
    });
  });

  it("handles incoming Wails events", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    // 1. Trigger ssh_disconnected event with correct and incorrect session IDs
    if (eventCallbacks["ssh_disconnected"]) {
      act(() => {
        eventCallbacks["ssh_disconnected"]("incorrect-id");
      });
      act(() => {
        eventCallbacks["ssh_disconnected"](mockSession.id);
      });
      expect(mockProps.addToast).toHaveBeenCalledWith("SSH", expect.stringContaining("Disconnected"), "warn");
    }

    // 2. Trigger ssh_path_changed event with correct/incorrect session IDs and duplicate paths
    if (eventCallbacks["ssh_path_changed"]) {
      act(() => {
        eventCallbacks["ssh_path_changed"]({ sessionId: "incorrect-id", path: "/home/user/other" });
      });
      act(() => {
        eventCallbacks["ssh_path_changed"]({ sessionId: mockSession.id, path: "/home/user" }); // duplicate path
      });
      act(() => {
        eventCallbacks["ssh_path_changed"]({ sessionId: mockSession.id, path: "/home/user/other" });
      });
    }

    // 3. Trigger ssh_output event with correct and incorrect session IDs
    if (eventCallbacks["ssh_output"]) {
      act(() => {
        eventCallbacks["ssh_output"]({ sessionId: "incorrect-id", data: "ignored output" });
        eventCallbacks["ssh_output"]({ sessionId: mockSession.id, data: "some output" });
      });
    }
  });

  it("handles Rename action and archive types", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("archive.zip")).toBeInTheDocument();
    });

    // 1. Right click on archive.zip
    const archiveItem = screen.getByText("archive.zip");
    fireEvent.contextMenu(archiveItem);

    // "Open With" should not be visible for archives
    expect(screen.queryByText("Open With")).not.toBeInTheDocument();

    // 2. Click Rename and perform rename successfully
    window.prompt = vi.fn().mockReturnValue("new_archive_name.zip");
    const renameBtn = screen.getByText("Rename");
    fireEvent.click(renameBtn);
    expect(AppBackend.ExecuteSFTPAction).toHaveBeenCalledWith(
      mockSession.id,
      "rename",
      "/home/user/archive.zip",
      "/home/user/new_archive_name.zip",
    );

    // 3. Test Rename with prompt cancellation (no input)
    fireEvent.contextMenu(archiveItem);
    window.prompt = vi.fn().mockReturnValue("");
    const renameBtn2 = screen.getByText("Rename");
    fireEvent.click(renameBtn2);
  });

  it("handles New Folder and Rename action failures", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    // Mock ExecuteSFTPAction to fail
    AppBackend.ExecuteSFTPAction.mockRejectedValueOnce(new Error("SFTP action failed"));

    // 1. New Folder failure
    window.prompt = vi.fn().mockReturnValue("fail_folder");
    const newBtn = screen.getByText("New");
    fireEvent.click(newBtn);
    await waitFor(() => {
      expect(mockProps.addToast).toHaveBeenCalledWith("Error", expect.stringContaining("SFTP action failed"), "error");
    });

    // 2. Rename failure
    const archiveItem = screen.getByText("archive.zip");
    fireEvent.contextMenu(archiveItem);

    AppBackend.ExecuteSFTPAction.mockRejectedValueOnce(new Error("Rename failed"));
    window.prompt = vi.fn().mockReturnValue("fail_rename.zip");
    const renameBtn = screen.getByText("Rename");
    fireEvent.click(renameBtn);
    await waitFor(() => {
      expect(mockProps.addToast).toHaveBeenCalledWith("Error", expect.stringContaining("Rename failed"), "error");
    });
  });

  it("handles Download and Upload action failures", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    const fileItem = screen.getByText("test.txt");
    fireEvent.contextMenu(fileItem);

    // 1. Download failure
    AppBackend.DownloadRemoteFile.mockRejectedValueOnce(new Error("Download failed"));
    const downloadBtn = screen.getByText("Download");
    fireEvent.click(downloadBtn);

    await waitFor(() => {
      expect(mockProps.addToast).toHaveBeenCalledWith("Error", expect.stringContaining("Download failed"), "error");
    });

    // 2. Upload failure
    AppBackend.UploadRemoteFile.mockRejectedValueOnce(new Error("Upload failed"));
    const uploadBtn = screen.getByText("Upload");
    fireEvent.click(uploadBtn);

    await waitFor(() => {
      expect(mockProps.addToast).toHaveBeenCalledWith("Error", expect.stringContaining("Upload failed"), "error");
    });
  });

  it("handles Edit and Delete action failures", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    // 1. Edit failure
    AppBackend.EditRemoteFile.mockRejectedValueOnce(new Error("Edit failed"));
    const fileItem = screen.getByText("test.txt");
    fireEvent.contextMenu(fileItem);
    const editBtn = screen.getByText("Edit File");
    fireEvent.click(editBtn);

    await waitFor(() => {
      expect(mockProps.addToast).toHaveBeenCalledWith("Error", expect.stringContaining("Edit failed"), "error");
    });

    // 2. Delete failure
    AppBackend.ExecuteSFTPAction.mockRejectedValueOnce(new Error("Delete failed"));
    fireEvent.contextMenu(fileItem);
    const deleteBtn = screen.getByText("Delete");
    fireEvent.click(deleteBtn);

    await waitFor(() => {
      expect(mockProps.addToast).toHaveBeenCalledWith("Error", expect.stringContaining("Delete failed"), "error");
    });
  });

  it("handles sorting and manual path entry navigation", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    // 1. Trigger sorting by name and size
    const nameHeader = screen.getByRole("button", { name: /Name/ });
    fireEvent.click(nameHeader);
    fireEvent.click(nameHeader); // toggle direction

    const sizeHeader = screen.getByRole("button", { name: /Size/ });
    fireEvent.click(sizeHeader);
    fireEvent.click(sizeHeader); // toggle direction

    // 2. Test manual navigation input
    // Find the input element containing current path "/home/user"
    const pathInput = screen.getByDisplayValue("/home/user");
    fireEvent.change(pathInput, { target: { value: "/home/user/manual" } });
    fireEvent.keyDown(pathInput, { key: "Enter", code: "Enter" });

    await waitFor(() => {
      expect(AppBackend.GetRemoteFiles).toHaveBeenCalledWith(mockSession.id, "/home/user/manual");
    });
  });

  it("handles navigation edge cases with trailing slashes", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    const pathInput = screen.getByDisplayValue("/home/user");
    fireEvent.change(pathInput, { target: { value: "/home/user/" } });
    fireEvent.keyDown(pathInput, { key: "Enter", code: "Enter" });

    await waitFor(() => {
      expect(AppBackend.GetRemoteFiles).toHaveBeenCalledWith(mockSession.id, "/home/user/");
    });

    const backBtn = screen.getByTitle("Back");
    fireEvent.click(backBtn);
  });

  it("handles terminal context menu actions", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    const terminalDiv = document.querySelector(".absolute.inset-0.px-2.pt-2");
    if (terminalDiv) {
      fireEvent.contextMenu(terminalDiv);

      // Verify menu rendering
      expect(screen.getByText("Copy")).toBeInTheDocument();
      expect(screen.getByText("Paste")).toBeInTheDocument();
      expect(screen.getByText("Refresh")).toBeInTheDocument();
      expect(screen.queryByText("Toggle view files/folder")).not.toBeInTheDocument();

      // 2. Click Copy
      const copyBtn = screen.getByText("Copy");
      fireEvent.click(copyBtn);
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith("selected terminal text");

      // 3. Click Paste
      fireEvent.contextMenu(terminalDiv);
      const pasteBtn = screen.getByText("Paste");
      fireEvent.click(pasteBtn);
      expect(navigator.clipboard.readText).toHaveBeenCalled();
      await waitFor(() => {
        expect(AppBackend.SendSSHInput).toHaveBeenCalledWith(mockSession.id, "pasted text");
      });

      // 4. Click Refresh
      fireEvent.contextMenu(terminalDiv);
      const refreshBtn = screen.getByText("Refresh");
      fireEvent.click(refreshBtn);

      // Open terminal context menu again near the bottom to test position adjustment
      const originalInnerHeight = window.innerHeight;
      Object.defineProperty(window, "innerHeight", {
        writable: true,
        configurable: true,
        value: 600,
      });

      fireEvent.contextMenu(terminalDiv, { clientY: 550, clientX: 200 });
      const termMenuElement = screen.getByText("Copy").closest("div");
      expect(termMenuElement).toBeInTheDocument();
      // Expected Y: 550 - 100 = 450
      expect(termMenuElement?.style.top).toBe("450px");

      // Restore innerHeight
      Object.defineProperty(window, "innerHeight", {
        writable: true,
        configurable: true,
        value: originalInnerHeight,
      });
    } else {
      expect(terminalDiv).not.toBeNull();
    }
  });

  it("suppresses all errors including access denied / permission denied during background/automatic sync", async () => {
    // 1. Mock GetRemoteFiles to fail with "Access is denied"
    vi.mocked(AppBackend.GetRemoteFiles).mockRejectedValueOnce(new Error("open \\\\wsl.localhost\\Ubuntu\\mnt\\d\\koding\\ostenia: Access is denied."));

    render(<SSHSessionView {...mockProps} />);

    // Since background/initial load uses isAutoSync=true, no toast error should be shown
    await waitFor(() => {
      expect(mockProps.addToast).not.toHaveBeenCalledWith("Explorer", expect.any(String), "error");
    });

    // 2. Mock GetRemoteFiles to fail with "Permission denied"
    vi.mocked(AppBackend.GetRemoteFiles).mockRejectedValueOnce(new Error("Permission denied"));

    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(mockProps.addToast).not.toHaveBeenCalledWith("Explorer", expect.any(String), "error");
    });

    // 4. Mock GetRemoteFiles to fail with "session not connected"
    vi.mocked(AppBackend.GetRemoteFiles).mockRejectedValueOnce(new Error("session not connected"));

    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(mockProps.addToast).not.toHaveBeenCalledWith("Explorer", expect.any(String), "error");
    });

    // 5. Mock GetRemoteFiles to fail with "sftp not connected"
    vi.mocked(AppBackend.GetRemoteFiles).mockRejectedValueOnce(new Error("sftp not connected"));

    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(mockProps.addToast).not.toHaveBeenCalledWith("Explorer", expect.any(String), "error");
    });

    // 6. Mock GetRemoteFiles to fail with "session not found"
    vi.mocked(AppBackend.GetRemoteFiles).mockRejectedValueOnce(new Error("session not found"));

    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(mockProps.addToast).not.toHaveBeenCalledWith("Explorer", expect.any(String), "error");
    });

    // 3. Mock GetRemoteFiles to fail with "EOF"
    vi.mocked(AppBackend.GetRemoteFiles).mockRejectedValueOnce(new Error("EOF"));

    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(mockProps.addToast).not.toHaveBeenCalledWith("Explorer", expect.any(String), "error");
    });
  });

  it("does not suppress errors during manual navigation or user-initiated refresh", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    // Mock GetRemoteFiles to fail with EOF on the next manual navigation call
    vi.mocked(AppBackend.GetRemoteFiles).mockRejectedValueOnce(new Error("connection EOF"));

    const pathInput = screen.getByDisplayValue("/home/user");
    fireEvent.change(pathInput, { target: { value: "/home/user/manual" } });
    fireEvent.keyDown(pathInput, { key: "Enter", code: "Enter" });

    // Since isManualEntry is true, it should show "Directory not available" as a "Navigation" toast, NOT suppress it
    await waitFor(() => {
      expect(mockProps.addToast).toHaveBeenCalledWith("Navigation", "Directory not available", "error");
    });
  });

  it("covers file double click, syncExplorer empty, and directory manual navigation edge cases", async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("test.txt")).toBeInTheDocument();
    });

    // 1. Double click file to edit
    const fileItem = screen.getByText("test.txt");
    fireEvent.doubleClick(fileItem);
    await waitFor(() => {
      expect(AppBackend.EditRemoteFile).toHaveBeenCalledWith(mockSession.id, "/home/user/test.txt");
    });

    // 2. Double click directory fallback when remotePath is empty
    // To test fallback when remotePath is empty, let's mock GetRemoteCurrentPath to return empty string
    vi.mocked(AppBackend.GetRemoteCurrentPath).mockResolvedValueOnce("");
    render(<SSHSessionView {...mockProps} />);

    const folderItem = await screen.findByText("folder");
    fireEvent.doubleClick(folderItem);

    // 3. syncExplorer when current path is empty (returns early)
    vi.mocked(AppBackend.GetRemoteCurrentPath).mockResolvedValueOnce("");
    const syncButtons = screen.getAllByTitle("Sync with terminal");
    if (syncButtons.length > 0) {
      fireEvent.click(syncButtons[0]);
    }

    // 4. syncExplorer with trailing slash path, e.g. "/home/user/" -> normalized to "/home/user"
    vi.mocked(AppBackend.GetRemoteCurrentPath).mockResolvedValueOnce("/home/user/");
    if (syncButtons.length > 0) {
      fireEvent.click(syncButtons[0]);
    }

    // 5. doubleClick directory when remotePath ends with slash
    vi.mocked(AppBackend.GetRemoteCurrentPath).mockResolvedValueOnce("/home/user/");
    render(<SSHSessionView {...mockProps} />);
    const folderItem2 = await screen.findByText("folder");
    fireEvent.doubleClick(folderItem2);

    // 6. navigateUp when path is "/" or empty (returns early)
    vi.mocked(AppBackend.GetRemoteCurrentPath).mockResolvedValueOnce("/");
    render(<SSHSessionView {...mockProps} />);
    const backBtn = screen.getAllByTitle("Back");
    if (backBtn.length > 0) {
      fireEvent.click(backBtn[0]);
    }
  });

  it("renders the real-time resource usage monitoring bar and toggles settings", async () => {
    vi.mocked(AppBackend.GetSSHResourceUsage).mockResolvedValue({
      cpu: 54,
      mem: 75,
      memTotal: 8192,
      memUsed: 6144,
      disk: 90,
      diskTotal: 102400,
      diskUsed: 92160,
    });

    render(<SSHSessionView {...mockProps} />);

    // Initially (before first fetch finishes) should show em-dash
    expect(screen.getAllByText(/—/).length).toBe(3);

    // Wait for the mock values to be loaded
    await waitFor(() => {
      expect(screen.getByText("CPU: 54%")).toBeInTheDocument();
    });

    expect(screen.getByText("RAM: 6.0 GB / 8.0 GB (75%)")).toBeInTheDocument();
    expect(screen.getByText("DISK: 90.0 GB / 100.0 GB (90%)")).toBeInTheDocument();

    // Hover over CPU item to trigger tooltip
    const cpuItem = screen.getByText("CPU: 54%");
    fireEvent.mouseEnter(cpuItem);
    expect(screen.getByText("CPU Usage History")).toBeInTheDocument();

    // Leave hover
    fireEvent.mouseLeave(cpuItem);
    expect(screen.queryByText("CPU Usage History")).not.toBeInTheDocument();

    // Verify gear settings button and click it to open settings
    const gearBtn = screen.getByTitle("Monitoring Settings");
    expect(gearBtn).toBeInTheDocument();
    fireEvent.click(gearBtn);

    expect(mockProps.onOpenSettings).toHaveBeenCalledWith("config");
  });

  it("handles connection drop and displays offline gray zone", async () => {
    // Mock first fetch to succeed
    vi.mocked(AppBackend.GetSSHResourceUsage).mockResolvedValueOnce({
      cpu: 50,
      mem: 60,
      memTotal: 8192,
      memUsed: 4915.2,
      disk: 70,
      diskTotal: 102400,
      diskUsed: 71680,
    });

    render(<SSHSessionView {...mockProps} />);

    // First fetch succeeds
    await waitFor(() => {
      expect(screen.getByText("CPU: 50%")).toBeInTheDocument();
    });
    expect(screen.getByText("RAM: 4.8 GB / 8.0 GB (60%)")).toBeInTheDocument();
    expect(screen.getByText("DISK: 70.0 GB / 100.0 GB (70%)")).toBeInTheDocument();

    // Now mock second fetch to fail
    vi.mocked(AppBackend.GetSSHResourceUsage).mockRejectedValue(new Error("Connection lost"));

    // Click Reconnect to trigger fresh connection and metric fetch
    const reconnectBtn = screen.getByTitle("Reconnect");
    fireEvent.click(reconnectBtn);

    await waitFor(() => {
      expect(screen.getByText("CPU: —")).toBeInTheDocument();
    });
    expect(screen.getByText("RAM: —")).toBeInTheDocument();
    expect(screen.getByText("DISK: —")).toBeInTheDocument();
  });
});
