import { render, screen, fireEvent, act } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import LogViewer from "./LogViewer";
import React from "react";

const generateManyLogs = (count: number) => {
  return Array.from({ length: count }, (_, idx) => ({
    id: String(idx + 1),
    time: `12:00:${idx.toString().padStart(2, '0')}`,
    msg: `Log entry number ${idx + 1}`,
  }));
};

describe("LogViewer Component", () => {
  const mockLogs = [
    {
      id: "1",
      time: "12:00:00",
      msg: "[SYS] Server starting",
      cleanMsg: "Server starting",
      type: "info",
      isServiceLog: false,
      caller: {
        functionName: "initServer",
        fileName: "App.tsx",
        line: "150",
        column: "4",
      },
      rawStack: "Error\n    at initServer (App.tsx:150:4)",
    },
    {
      id: "2",
      time: "12:00:05",
      msg: "[WRN] Low disk space",
      cleanMsg: "Low disk space",
      type: "warn",
      isServiceLog: false,
      caller: {
        functionName: "checkDisk",
        fileName: "diskUtils.ts",
        line: "45",
        column: "12",
      },
      rawStack: "Error\n    at checkDisk (diskUtils.ts:45:12)",
    },
    {
      id: "3",
      time: "12:00:10",
      msg: "[ERR] Database connection failed",
      cleanMsg: "Database connection failed",
      type: "error",
      isServiceLog: false,
      caller: {
        functionName: "connectDb",
        fileName: "db.ts",
        line: "99",
        column: "8",
      },
      rawStack: "Error\n    at connectDb (db.ts:99:8)",
    },
    {
      id: "4",
      time: "12:00:15",
      msg: "[Apache] AH00558: apache2: Could not reliably determine the server's fully qualified domain name",
      cleanMsg: "AH00558: apache2: Could not reliably determine the server's fully qualified domain name",
      type: "info",
      isServiceLog: true,
    },
  ];

  beforeEach(() => {
    vi.useFakeTimers();
    // Mock navigator.clipboard
    Object.defineProperty(navigator, "clipboard", {
      value: {
        writeText: vi.fn().mockResolvedValue(undefined),
      },
      configurable: true,
      writable: true,
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it("renders empty state when no logs are provided", () => {
    render(<LogViewer logs={[]} />);
    expect(screen.getByText(/No activity recorded yet/i)).toBeInTheDocument();
  });

  it("renders simple log mode by default and allows toggling to complete log mode", () => {
    render(<LogViewer logs={mockLogs} />);

    // Default should be simple mode
    expect(screen.getByText("[SYS] Server starting")).toBeInTheDocument();
    // In simple mode, caller info shouldn't be visible
    expect(screen.queryByText("initServer()")).not.toBeInTheDocument();

    // Toggle to Complete Log
    const completeButton = screen.getByRole("button", { name: /complete log/i });
    fireEvent.click(completeButton);

    // In complete mode, we should see badges
    expect(screen.getByText("INFO")).toBeInTheDocument();
    expect(screen.getByText("WARN")).toBeInTheDocument();
    expect(screen.getByText("ERROR")).toBeInTheDocument();
    expect(screen.getByText("SERVICE")).toBeInTheDocument();

    // Caller info should be visible for non-service logs
    expect(screen.getByText("initServer()")).toBeInTheDocument();
    expect(screen.getByText("App.tsx:150:4")).toBeInTheDocument();
    expect(screen.getByText("checkDisk()")).toBeInTheDocument();
    expect(screen.getByText("diskUtils.ts:45:12")).toBeInTheDocument();

    // No caller info for service logs
    const serviceCard = screen.getByText(/AH00558/);
    expect(serviceCard).toBeInTheDocument();

    // Switch back to Simple Log
    const simpleButton = screen.getByRole("button", { name: /simple log/i });
    fireEvent.click(simpleButton);
    expect(screen.queryByText("initServer()")).not.toBeInTheDocument();
  });

  it("displays expandable stack trace inside complete log cards", () => {
    render(<LogViewer logs={mockLogs} />);

    // Switch to complete log mode
    const completeButton = screen.getByRole("button", { name: /complete log/i });
    fireEvent.click(completeButton);

    // Verify "View Stack Trace" elements
    const viewStackTraces = screen.getAllByText("View Stack Trace");
    expect(viewStackTraces).toHaveLength(3); // 3 non-service logs have stacks

    // Verify raw stack trace exists inside <pre>
    const preBlocks = screen.getAllByText(/at initServer/i);
    expect(preBlocks.length).toBeGreaterThan(0);
  });

  it("successfully copies structured Laravel-style format to clipboard in simple mode", async () => {
    render(<LogViewer logs={[mockLogs[0]]} />);

    // Find the copy button
    const copyButtons = screen.getAllByTitle("Copy Log");
    expect(copyButtons).toHaveLength(1);

    // Click copy button
    fireEvent.click(copyButtons[0]);
    await act(async () => {});

    // Check navigator.clipboard.writeText was called with expected Laravel-style structure
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("Timestamp: [12:00:00]")
    );
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("Level: INFO")
    );
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("Message: Server starting")
    );
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("Caller Function: initServer()")
    );
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("File: App.tsx:150:4")
    );
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("Stack Trace:\nError\n    at initServer (App.tsx:150:4)")
    );

    // Verify the copy button has visual feedback (check icon)
    expect(screen.getByTitle("Copy Log")).toHaveClass("text-emerald-500");

    // Advance time by 2 seconds
    act(() => {
      vi.advanceTimersByTime(2000);
    });

    // Verify icon / style resets
    expect(screen.getByTitle("Copy Log")).not.toHaveClass("text-emerald-500");
  });

  it("successfully copies service log without caller info to clipboard in complete mode", async () => {
    render(<LogViewer logs={[mockLogs[3]]} />);

    // Switch to complete log mode
    const completeButton = screen.getByRole("button", { name: /complete log/i });
    fireEvent.click(completeButton);

    // Find the copy button
    const copyButton = screen.getByTitle("Copy Laravel-Style Log");
    expect(copyButton).toBeInTheDocument();

    // Click copy button
    fireEvent.click(copyButton);
    await act(async () => {});

    // Check navigator.clipboard.writeText was called with expected Laravel-style structure for Service log
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("Timestamp: [12:00:15]")
    );
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("Level: SERVICE")
    );
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("Caller Function: N/A")
    );
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("File: N/A")
    );
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("Stack Trace:\nN/A")
    );
  });

  describe("Gmail-style Selection Mode, Shift+Click, Context Menu, and Batch Copy", () => {
    it("toggles selection mode, shows checkboxes, and handles single/select-all selection", () => {
      render(<LogViewer logs={mockLogs} />);

      // Initially "Select Log" button should be present and no checkboxes
      const selectLogBtn = screen.getByRole("button", { name: /select log/i });
      expect(selectLogBtn).toBeInTheDocument();
      expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();

      // Click "Select Log" to enter selection mode
      fireEvent.click(selectLogBtn);

      // Should now see "Cancel Selection" button, "Select All (Current Page)" checkbox, and 4 entry checkboxes
      expect(screen.getByRole("button", { name: /cancel selection/i })).toBeInTheDocument();
      const checkboxes = screen.getAllByRole("checkbox");
      // 1 header checkbox + 4 log checkboxes = 5 total
      expect(checkboxes).toHaveLength(5);

      // Check "Select All (Current Page)"
      const selectAllCheckbox = screen.getByLabelText(/select all \(current page\)/i);
      fireEvent.click(selectAllCheckbox);

      // Count of selected items should update
      expect(screen.getByText(/4 selected/i)).toBeInTheDocument();

      // Uncheck "Select All"
      fireEvent.click(selectAllCheckbox);
      expect(screen.getByText(/0 selected/i)).toBeInTheDocument();

      // Click "Cancel Selection" to exit selection mode
      const cancelBtn = screen.getByRole("button", { name: /cancel selection/i });
      fireEvent.click(cancelBtn);

      expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
      expect(screen.getByRole("button", { name: /select log/i })).toBeInTheDocument();
    });

    it("opens context menu with 'Select Log' when right clicked outside selection mode", async () => {
      render(<LogViewer logs={mockLogs} />);

      const logEntry = screen.getByText("[12:00:00]");
      fireEvent.contextMenu(logEntry);

      // Menu option inside context menu popover is the first 'Select Log' button in DOM
      const selectLogOptions = screen.getAllByRole("button", { name: /select log/i });
      const contextMenuOption = selectLogOptions[0];

      await act(async () => {
        fireEvent.click(contextMenuOption);
      });

      // Selection mode should be activated and target log selected
      expect(screen.getByRole("button", { name: /cancel selection/i })).toBeInTheDocument();
      expect(screen.getByText(/1 selected/i)).toBeInTheDocument();
    });

    it("supports Shift+Click range selection like Gmail", () => {
      render(<LogViewer logs={mockLogs} />);

      // Enter selection mode
      fireEvent.click(screen.getByRole("button", { name: /select log/i }));

      const checkboxes = screen.getAllByRole("checkbox");

      // Click first item checkbox (index 1)
      fireEvent.click(checkboxes[1]);
      expect(screen.getByText(/1 selected/i)).toBeInTheDocument();

      // Shift+Click fourth item checkbox (index 4)
      fireEvent.click(checkboxes[4], { shiftKey: true });

      // Range 1 to 4 should be selected
      expect(screen.getByText(/4 selected/i)).toBeInTheDocument();
    });

    it("copies selected logs with Copied! tooltip feedback", async () => {
      render(<LogViewer logs={mockLogs} />);

      // Enter selection mode
      fireEvent.click(screen.getByRole("button", { name: /select log/i }));

      // Select first two logs
      const checkboxes = screen.getAllByRole("checkbox");
      fireEvent.click(checkboxes[1]); // log 1
      fireEvent.click(checkboxes[2]); // log 2

      expect(screen.getByText(/2 selected/i)).toBeInTheDocument();

      // Click Copy button in action bar
      const copyBtn = screen.getByRole("button", { name: /^copy$/i });
      fireEvent.click(copyBtn);
      await act(async () => {});

      // Verify exact WhatsApp-style text format [time] msg
      const expectedText = "[12:00:00] [SYS] Server starting\n[12:00:05] [WRN] Low disk space";
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith(expectedText);

      // Verify Copied! tooltip feedback appeared
      expect(screen.getByText("Copied!")).toBeInTheDocument();

      // Advance time by 2 seconds
      act(() => {
        vi.advanceTimersByTime(2000);
      });

      expect(screen.queryByText("Copied!")).not.toBeInTheDocument();
    });

    it("opens context menu on right click in selection mode and handles copy/cancel", async () => {
      render(<LogViewer logs={mockLogs} />);

      // Enter selection mode
      fireEvent.click(screen.getByRole("button", { name: /select log/i }));

      // Right-click on third log entry
      const logEntry = screen.getByText("[12:00:10]");
      fireEvent.contextMenu(logEntry);

      // Context menu should appear with Copy (1) and Cancel
      const copyMenuBtn = screen.getByRole("button", { name: /copy \(1\)/i });

      expect(copyMenuBtn).toBeInTheDocument();

      // Click Copy in context menu
      fireEvent.click(copyMenuBtn);
      await act(async () => {});

      expect(navigator.clipboard.writeText).toHaveBeenCalledWith("[12:00:10] [ERR] Database connection failed");
    });
  });

  describe("Pagination and Ordering", () => {
    it("respects log list page limits and supports next/prev pagination", () => {
      const manyLogs = generateManyLogs(25);

      render(<LogViewer logs={manyLogs} />);

      // Verify that pagination controls are present and showing correct count
      const showingText = screen.getByText((content, element) => element?.textContent === "Showing 1 to 20 of 25 entries");
      expect(showingText).toBeInTheDocument();

      // Verify that first 20 items are visible, but the 21st is not
      expect(screen.getByText("Log entry number 1")).toBeInTheDocument();
      expect(screen.getByText("Log entry number 20")).toBeInTheDocument();
      expect(screen.queryByText("Log entry number 21")).not.toBeInTheDocument();

      // Click "Next" to go to page 2
      const nextButton = screen.getByRole("button", { name: /next/i });
      fireEvent.click(nextButton);

      // Verify page details and showing count
      const showingPage2Text = screen.getByText((content, element) => element?.textContent === "Showing 21 to 25 of 25 entries");
      expect(showingPage2Text).toBeInTheDocument();

      const page2Button = screen.getByRole("button", { name: "2" });
      expect(page2Button).toBeInTheDocument();
      expect(page2Button).toHaveClass("bg-blue-600");

      // On page 2, verify item 21 is visible, but item 20 is gone
      expect(screen.getByText("Log entry number 21")).toBeInTheDocument();
      expect(screen.queryByText("Log entry number 20")).not.toBeInTheDocument();

      // Click "Previous" to return to page 1
      const prevButton = screen.getByRole("button", { name: /previous/i });
      fireEvent.click(prevButton);

      const showingPage1Text = screen.getByText((content, element) => element?.textContent === "Showing 1 to 20 of 25 entries");
      expect(showingPage1Text).toBeInTheDocument();
    });

    it("supports page size dropdown changes", () => {
      const manyLogs = generateManyLogs(15);

      render(<LogViewer logs={manyLogs} />);

      // Default page size is 20, so all 15 logs are shown
      const showingText = screen.getByText((content, element) => element?.textContent === "Showing 1 to 15 of 15 entries");
      expect(showingText).toBeInTheDocument();

      // Change page size to 10
      const select = screen.getByRole("combobox");
      fireEvent.change(select, { target: { value: "10" } });

      // Verify limit updated to 10
      const showingUpdatedText = screen.getByText((content, element) => element?.textContent === "Showing 1 to 10 of 15 entries");
      expect(showingUpdatedText).toBeInTheDocument();
      expect(screen.getByText("Log entry number 1")).toBeInTheDocument();
      expect(screen.queryByText("Log entry number 11")).not.toBeInTheDocument();
    });
  });

  describe("Real-time Log Buffering", () => {
    it("does not buffer logs and updates immediately when isActive is false", () => {
      const initialLogs = [
        { id: "1", time: "12:00:00", msg: "Log 1" },
        { id: "2", time: "12:00:01", msg: "Log 2" },
      ];
      const { rerender } = render(<LogViewer logs={initialLogs} isActive={false} />);

      expect(screen.getByText("Log 1")).toBeInTheDocument();
      expect(screen.getByText("Log 2")).toBeInTheDocument();
      expect(screen.queryByText(/new logs/i)).not.toBeInTheDocument();

      const updatedLogs = [
        { id: "3", time: "12:00:02", msg: "Log 3" },
        ...initialLogs,
      ];
      rerender(<LogViewer logs={updatedLogs} isActive={false} />);

      expect(screen.getByText("Log 3")).toBeInTheDocument();
      expect(screen.queryByText(/new logs/i)).not.toBeInTheDocument();
    });

    it("buffers new logs and shows a floating badge when isActive is true", () => {
      const initialLogs = [
        { id: "1", time: "12:00:00", msg: "Log 1" },
        { id: "2", time: "12:00:01", msg: "Log 2" },
      ];
      const { rerender } = render(<LogViewer logs={initialLogs} isActive={true} />);

      expect(screen.getByText("Log 1")).toBeInTheDocument();
      expect(screen.getByText("Log 2")).toBeInTheDocument();
      expect(screen.queryByText(/new logs/i)).not.toBeInTheDocument();

      // Add a third log
      const updatedLogs = [
        { id: "3", time: "12:00:02", msg: "Log 3" },
        ...initialLogs,
      ];
      rerender(<LogViewer logs={updatedLogs} isActive={true} />);

      // Log 3 should NOT be in the document since it is buffered (pending)
      expect(screen.queryByText("Log 3")).not.toBeInTheDocument();

      // A floating badge should represent the pending logs
      const badge = screen.getByText("1 new logs");
      expect(badge).toBeInTheDocument();

      // Test hover behavior
      fireEvent.mouseEnter(badge);
      expect(screen.getByText("show 1 new logs")).toBeInTheDocument();

      fireEvent.mouseLeave(badge);
      expect(screen.getByText("1 new logs")).toBeInTheDocument();

      // Click the badge to release the buffered logs
      fireEvent.click(badge);

      // Now Log 3 should be displayed
      expect(screen.getByText("Log 3")).toBeInTheDocument();
      expect(screen.queryByText(/new logs/i)).not.toBeInTheDocument();
    });

    it("discards or handles completely new log lists (like a reset or tab change reload)", () => {
      const initialLogs = [
        { id: "1", time: "12:00:00", msg: "Log 1" },
      ];
      const { rerender } = render(<LogViewer logs={initialLogs} isActive={true} />);

      expect(screen.getByText("Log 1")).toBeInTheDocument();

      // A completely different list of logs (e.g., cleared/re-fetched logs)
      const clearedLogs: any[] = [];
      rerender(<LogViewer logs={clearedLogs} isActive={true} />);
      expect(screen.queryByText("Log 1")).not.toBeInTheDocument();
      expect(screen.getByText(/No activity recorded yet/i)).toBeInTheDocument();
    });
  });
});
