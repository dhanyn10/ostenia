import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import AppHeader from "./AppHeader";
import React from "react";

describe("AppHeader Component", () => {
  it("renders correctly for different tabs", () => {
    const { rerender } = render(
      <AppHeader
        activeTab="activity"
        handleStartAll={() => {}}
        handleStopAll={() => {}}
        handleTerminal={() => {}}
        isTerminalOpen={false}
        setIsTerminalOpen={() => {}}
      />,
    );
    expect(screen.getByText("Activity Center")).toBeInTheDocument();

    rerender(
      <AppHeader
        activeTab="plugins"
        handleStartAll={() => {}}
        handleStopAll={() => {}}
        handleTerminal={() => {}}
        isTerminalOpen={false}
        setIsTerminalOpen={() => {}}
      />,
    );
    expect(screen.getByText("Plugin Management")).toBeInTheDocument();

    rerender(
      <AppHeader
        activeTab="proxy"
        handleStartAll={() => {}}
        handleStopAll={() => {}}
        handleTerminal={() => {}}
        isTerminalOpen={false}
        setIsTerminalOpen={() => {}}
      />,
    );
    expect(screen.getByText("Proxy Management")).toBeInTheDocument();

    rerender(
      <AppHeader
        activeTab="ssh"
        handleStartAll={() => {}}
        handleStopAll={() => {}}
        handleTerminal={() => {}}
        isTerminalOpen={false}
        setIsTerminalOpen={() => {}}
      />,
    );
    expect(screen.queryByText("SSH & Remote Files")).not.toBeInTheDocument();
  });

  it("returns null when activeTab is logs", () => {
    const { container } = render(
      <AppHeader
        activeTab="logs"
        handleStartAll={() => {}}
        handleStopAll={() => {}}
        handleTerminal={() => {}}
        isTerminalOpen={false}
        setIsTerminalOpen={() => {}}
      />,
    );
    expect(container.firstChild).toBeNull();
  });

  it("handles start and stop all buttons", () => {
    const handleStartAll = vi.fn();
    const handleStopAll = vi.fn();
    render(
      <AppHeader
        activeTab="activity"
        handleStartAll={handleStartAll}
        handleStopAll={handleStopAll}
        handleTerminal={() => {}}
        isTerminalOpen={false}
        setIsTerminalOpen={() => {}}
      />,
    );

    fireEvent.click(screen.getByText(/Start All/i));
    expect(handleStartAll).toHaveBeenCalledTimes(1);

    fireEvent.click(screen.getByText(/Stop All/i));
    expect(handleStopAll).toHaveBeenCalledTimes(1);
  });

  it("toggles terminal dropdown", () => {
    const setIsTerminalOpen = vi.fn();
    const { rerender } = render(
      <AppHeader
        activeTab="activity"
        handleStartAll={() => {}}
        handleStopAll={() => {}}
        handleTerminal={() => {}}
        isTerminalOpen={false}
        setIsTerminalOpen={setIsTerminalOpen}
      />,
    );

    const buttons = screen.getAllByRole("button");
    // Start All, Stop All, and the toggle button
    const toggleBtn = buttons.find((btn) => btn.className.includes("p-2"));

    if (toggleBtn) {
      fireEvent.click(toggleBtn);
      expect(setIsTerminalOpen).toHaveBeenCalledWith(true);
    }

    rerender(
      <AppHeader
        activeTab="activity"
        handleStartAll={() => {}}
        handleStopAll={() => {}}
        handleTerminal={() => {}}
        isTerminalOpen={true}
        setIsTerminalOpen={setIsTerminalOpen}
      />,
    );
    expect(screen.getByText("Command Prompt (CMD)")).toBeInTheDocument();
  });

  it("calls handleTerminal with correct arguments", () => {
    const handleTerminal = vi.fn();
    render(
      <AppHeader
        activeTab="activity"
        handleStartAll={() => {}}
        handleStopAll={() => {}}
        handleTerminal={handleTerminal}
        isTerminalOpen={true}
        setIsTerminalOpen={() => {}}
      />,
    );

    fireEvent.click(screen.getByText("Command Prompt (CMD)"));
    expect(handleTerminal).toHaveBeenCalledWith("cmd");

    fireEvent.click(screen.getByText("PowerShell"));
    expect(handleTerminal).toHaveBeenCalledWith("powershell");

    fireEvent.click(screen.getByText("Git Bash"));
    expect(handleTerminal).toHaveBeenCalledWith("gitbash");
  });
});
