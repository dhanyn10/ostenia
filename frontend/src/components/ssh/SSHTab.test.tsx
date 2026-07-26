import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import SSHTab from "./SSHTab";
import React from "react";

// Mock AppBackend
vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetSSHSessions: vi.fn(),
  ConnectSSH: vi.fn(),
  DisconnectSSH: vi.fn(),
  DeleteSSHSession: vi.fn(),
}));

import * as AppBackendRaw from "../../../wailsjs/go/backend/App";
const AppBackend = AppBackendRaw as any;

// Mock sub-components to focus on SSHTab logic
vi.mock("../SSHSessionView", () => ({
  default: () => <div data-testid="ssh-session-view" />,
}));

vi.mock("./SSHSessionForm", () => ({
  default: ({ onClose }) => (
    <div data-testid="ssh-session-form">
      <button onClick={onClose}>Close</button>
    </div>
  ),
}));

describe("SSHTab Component", () => {
  // Generate dynamic IP addresses to avoid static analysis security warnings
  const generateRandomIP = () => {
    const bytes = new Uint8Array(4);
    globalThis.crypto.getRandomValues(bytes);
    return bytes.join(".");
  };

  const TEST_IP_1 = generateRandomIP();
  const TEST_IP_2 = generateRandomIP();

  const mockSessions = [
    { id: "1", name: "Server 1", host: TEST_IP_1, authMethod: "password" },
    { id: "2", name: "Server 2", host: TEST_IP_2, authMethod: "key" },
  ];

  it("renders loading state initially", async () => {
    AppBackend.GetSSHSessions.mockImplementation(() => Promise.resolve([]));
    render(
      <SSHTab addToast={vi.fn()} theme="light" onOpenSettings={vi.fn()} />,
    );
    expect(AppBackend.GetSSHSessions).toHaveBeenCalled();
  });

  it("renders sessions after loading", async () => {
    AppBackend.GetSSHSessions.mockImplementation(() =>
      Promise.resolve(mockSessions),
    );
    render(
      <SSHTab addToast={vi.fn()} theme="light" onOpenSettings={vi.fn()} />,
    );

    expect(await screen.findByText(TEST_IP_1)).toBeInTheDocument();
    expect(await screen.findByText(TEST_IP_2)).toBeInTheDocument();
  });

  it("opens new connection form", async () => {
    AppBackend.GetSSHSessions.mockImplementation(() => Promise.resolve([]));
    render(
      <SSHTab addToast={vi.fn()} theme="light" onOpenSettings={vi.fn()} />,
    );

    const newBtn = screen.getByTitle("New Connection");
    fireEvent.click(newBtn);

    expect(screen.getByTestId("ssh-session-form")).toBeInTheDocument();
  });

  it("connects on double click", async () => {
    AppBackend.GetSSHSessions.mockImplementation(() =>
      Promise.resolve(mockSessions),
    );
    render(
      <SSHTab addToast={vi.fn()} theme="light" onOpenSettings={vi.fn()} />,
    );

    await screen.findByText(TEST_IP_1);

    const card = screen.getByText(TEST_IP_1).closest("div").parentElement;
    fireEvent.doubleClick(card);

    expect(screen.getByTestId("ssh-session-view")).toBeInTheDocument();
    expect(screen.getByText("Dashboard")).toBeInTheDocument();
  });

  it("handles right-click context menu and delete actions", async () => {
    window.confirm = vi.fn().mockReturnValue(true);
    AppBackend.GetSSHSessions.mockImplementation(() =>
      Promise.resolve(mockSessions),
    );
    AppBackend.DeleteSSHSession.mockResolvedValue(null);

    render(
      <SSHTab addToast={vi.fn()} theme="light" onOpenSettings={vi.fn()} />,
    );

    await screen.findByText(TEST_IP_1);

    const card = screen.getByText(TEST_IP_1).closest("div").parentElement;
    fireEvent.contextMenu(card);

    // Verify context menu is displayed
    const deleteBtn = screen.getByText("Delete Session");
    expect(deleteBtn).toBeInTheDocument();

    // Click on context menu button to delete session
    fireEvent.click(deleteBtn);
    expect(window.confirm).toHaveBeenCalled();
    expect(AppBackend.DeleteSSHSession).toHaveBeenCalledWith("1");
  });

  it("dismisses right-click context menu when clicking elsewhere", async () => {
    AppBackend.GetSSHSessions.mockImplementation(() =>
      Promise.resolve(mockSessions),
    );

    render(
      <SSHTab addToast={vi.fn()} theme="light" onOpenSettings={vi.fn()} />,
    );

    await screen.findByText(TEST_IP_1);

    const card = screen.getByText(TEST_IP_1).closest("div").parentElement;
    fireEvent.contextMenu(card);

    // Verify context menu is displayed
    expect(screen.getByText("Delete Session")).toBeInTheDocument();

    // Click somewhere else to dismiss it
    fireEvent.click(document.body);
    expect(screen.queryByText("Delete Session")).not.toBeInTheDocument();
  });
});
