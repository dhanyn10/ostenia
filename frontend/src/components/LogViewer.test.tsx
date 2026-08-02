import { render, screen, fireEvent, act } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import LogViewer from "./LogViewer";
import React from "react";

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
    expect(viewStackTraces.length).toBe(3); // 3 non-service logs have stacks

    // Verify raw stack trace exists inside <pre>
    const preBlocks = screen.getAllByText(/at initServer/i);
    expect(preBlocks.length).toBeGreaterThan(0);
  });

  it("successfully copies structured Laravel-style format to clipboard in simple mode", async () => {
    render(<LogViewer logs={[mockLogs[0]]} />);

    // Find the copy button
    const copyButtons = screen.getAllByTitle("Copy Log");
    expect(copyButtons.length).toBe(1);

    // Click copy button inside act
    await act(async () => {
      fireEvent.click(copyButtons[0]);
    });

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
    await act(async () => {
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
    await act(async () => {
      fireEvent.click(copyButton);
    });

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
});
