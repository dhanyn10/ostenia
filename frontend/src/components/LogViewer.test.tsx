import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import LogViewer from "./LogViewer";
import React from "react";

describe("LogViewer Component", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Mock navigator.clipboard
    Object.defineProperty(navigator, "clipboard", {
      value: {
        writeText: vi.fn().mockResolvedValue(undefined),
      },
      writable: true,
    });
  });

  it("renders empty state when no logs are provided", () => {
    render(<LogViewer logs={[]} />);
    expect(screen.getByText(/No activity recorded yet/i)).toBeInTheDocument();
  });

  it("renders logs correctly", () => {
    const logs = [
      { id: "1", time: "10:00:00", msg: "Normal message" },
      { id: "2", time: "10:00:01", msg: "Error: something failed" },
      { id: "3", time: "10:00:02", msg: "Ready success" },
      { id: "4", time: "10:00:03", msg: "[WRN] warning" },
    ];
    render(<LogViewer logs={logs} />);

    expect(screen.getByText("Normal message")).toBeInTheDocument();
    expect(screen.getByText("Error: something failed")).toBeInTheDocument();
    expect(screen.getByText("Ready success")).toBeInTheDocument();
    expect(screen.getByText("[WRN] warning")).toBeInTheDocument();
  });

  it("applies correct color classes based on log content", () => {
    const logs = [
      { id: "1", time: "10:00:01", msg: "Error: something failed" },
      { id: "2", time: "10:00:02", msg: "Ready success" },
      { id: "3", time: "10:00:03", msg: "[WRN] warning" },
    ];
    render(<LogViewer logs={logs} />);

    const errorLog = screen.getByText("Error: something failed");
    expect(errorLog).toHaveClass("text-rose-500");

    const successLog = screen.getByText("Ready success");
    expect(successLog).toHaveClass("text-emerald-500");

    const warningLog = screen.getByText("[WRN] warning");
    expect(warningLog).toHaveClass("text-amber-500");
  });

  it("toggles between Simple Log and Complete Log modes", async () => {
    const logs = [
      {
        id: "1",
        time: "10:00:01",
        msg: "[ERR] Error: something failed",
        rawMsg: "Error: something failed",
        type: "error",
        caller: {
          functionName: "handleToggleService",
          fileName: "App.tsx",
          line: "392",
          column: "5",
          rawStack: "Error\n  at handleToggleService in App.tsx:392:5",
        },
      },
    ];

    render(<LogViewer logs={logs} />);

    // Starts in Simple Log mode by default
    expect(screen.getByText("[ERR] Error: something failed")).toBeInTheDocument();
    expect(screen.queryByText("Caller Function:")).not.toBeInTheDocument();

    // Click Complete Log button
    const completeBtn = screen.getByTestId("complete-log-btn");
    fireEvent.click(completeBtn);

    // Now in Complete Log mode (Laravel style card with details)
    expect(screen.getByText("Caller Function:")).toBeInTheDocument();
    expect(screen.getByText("handleToggleService()")).toBeInTheDocument();
    expect(screen.getByText("App.tsx:392:5")).toBeInTheDocument();

    // Toggle back to Simple Log mode
    const simpleBtn = screen.getByTestId("simple-log-btn");
    fireEvent.click(simpleBtn);

    expect(screen.queryByText("Caller Function:")).not.toBeInTheDocument();
  });

  it("copies complete log format to clipboard when copy button is clicked", async () => {
    const logs = [
      {
        id: "log-123",
        time: "10:00:01",
        msg: "[ERR] Error: something failed",
        rawMsg: "Error: something failed",
        type: "error",
        caller: {
          functionName: "handleToggleService",
          fileName: "App.tsx",
          line: "392",
          column: "5",
          rawStack: "Error\n  at handleToggleService in App.tsx:392:5",
        },
      },
    ];

    render(<LogViewer logs={logs} />);

    // Hover copy button is in simple log view
    const copyBtn = screen.getByTitle("Copy log details");
    expect(copyBtn).toBeInTheDocument();

    fireEvent.click(copyBtn);

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("Timestamp: [10:00:01]")
    );
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("Level: ERROR")
    );
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("Caller Function: handleToggleService()")
    );
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining("File: App.tsx:392:5")
    );
  });
});
