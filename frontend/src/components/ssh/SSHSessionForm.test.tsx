import {
  render,
  screen,
  fireEvent,
  waitFor,
  act,
} from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import React from "react";
import SSHSessionForm from "./SSHSessionForm";

// Mock AppBackend
vi.mock("../../../wailsjs/go/backend/App", () => ({
  AddSSHSession: vi.fn(),
  UpdateSSHSession: vi.fn(),
  GetWSLDistros: vi.fn().mockResolvedValue([]),
}));

import * as AppBackendRaw from "../../../wailsjs/go/backend/App";
const AppBackend = AppBackendRaw as any;

describe("SSHSessionForm Component", () => {
  const mockSession = {
    id: "session-123",
    name: "Production Server",
    host: "192.168.1.100",
    port: 22,
    user: "ubuntu",
    authMethod: "password",
    password: "mypassword123",
    keyPath: "",
    passphrase: "",
    createdAt: 1234567890,
  };

  const addToastMock = vi.fn();
  const onCloseMock = vi.fn();
  const onSaveMock = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders in New Connection mode by default", () => {
    render(
      <SSHSessionForm
        session={null}
        onClose={onCloseMock}
        onSave={onSaveMock}
        addToast={addToastMock}
      />,
    );

    expect(screen.getByText("New Connection")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("e.g. Production Web")).toHaveValue("");
    expect(screen.getByPlaceholderText("1.2.3.4 or example.com")).toHaveValue(
      "",
    );
    expect(screen.getByLabelText("Port")).toHaveValue(22);
    expect(screen.getByPlaceholderText("root")).toHaveValue("root");
  });

  it("renders with loaded session data in Edit Connection mode", () => {
    render(
      <SSHSessionForm
        session={mockSession}
        onClose={onCloseMock}
        onSave={onSaveMock}
        addToast={addToastMock}
      />,
    );

    expect(screen.getByText("Edit Connection")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("e.g. Production Web")).toHaveValue(
      "Production Server",
    );
    expect(screen.getByPlaceholderText("1.2.3.4 or example.com")).toHaveValue(
      "192.168.1.100",
    );
    expect(screen.getByLabelText("Port")).toHaveValue(22);
    expect(screen.getByPlaceholderText("root")).toHaveValue("ubuntu");
    expect(screen.getByPlaceholderText("••••••••")).toHaveValue(
      "mypassword123",
    );
  });

  it("allows toggling auth methods between Password and Key File", async () => {
    render(
      <SSHSessionForm
        session={null}
        onClose={onCloseMock}
        onSave={onSaveMock}
        addToast={addToastMock}
      />,
    );

    // Toggle to Key File
    const keyFileBtn = screen.getByRole("button", { name: "Key File" });
    fireEvent.click(keyFileBtn);

    expect(
      screen.getByPlaceholderText("e.g. /home/user/.ssh/id_rsa"),
    ).toBeInTheDocument();
    expect(screen.getByPlaceholderText("optional")).toBeInTheDocument(); // passphrase field

    // Toggle back to Password
    const passwordBtn = screen.getByRole("button", { name: "Password" });
    fireEvent.click(passwordBtn);

    expect(screen.getByPlaceholderText("••••••••")).toBeInTheDocument();
  });

  it("submits a new connection successfully", async () => {
    AppBackend.AddSSHSession.mockResolvedValue(true);

    render(
      <SSHSessionForm
        session={null}
        onClose={onCloseMock}
        onSave={onSaveMock}
        addToast={addToastMock}
      />,
    );

    fireEvent.change(screen.getByPlaceholderText("e.g. Production Web"), {
      target: { value: "My Test Server" },
    });
    fireEvent.change(screen.getByPlaceholderText("1.2.3.4 or example.com"), {
      target: { value: "10.0.0.1" },
    });
    fireEvent.change(screen.getByLabelText("Port"), {
      target: { value: "2222" },
    });
    fireEvent.change(screen.getByPlaceholderText("root"), {
      target: { value: "admin" },
    });
    fireEvent.change(screen.getByPlaceholderText("••••••••"), {
      target: { value: "secret" },
    });

    const saveBtn = screen.getByRole("button", { name: "Save" });
    fireEvent.click(saveBtn);

    expect(AppBackend.AddSSHSession).toHaveBeenCalled();
    await waitFor(() => {
      expect(addToastMock).toHaveBeenCalledWith(
        "Success",
        expect.stringContaining("created successfully"),
        "success",
      );
    });
    expect(onSaveMock).toHaveBeenCalled();
  });

  it("submits update successfully for an existing connection", async () => {
    AppBackend.UpdateSSHSession.mockResolvedValue(true);

    render(
      <SSHSessionForm
        session={mockSession}
        onClose={onCloseMock}
        onSave={onSaveMock}
        addToast={addToastMock}
      />,
    );

    const updateBtn = screen.getByRole("button", { name: "Update" });
    fireEvent.click(updateBtn);

    expect(AppBackend.UpdateSSHSession).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "session-123",
        name: "Production Server",
      }),
    );
    await waitFor(() => {
      expect(addToastMock).toHaveBeenCalledWith(
        "Success",
        expect.stringContaining("updated successfully"),
        "success",
      );
    });
    expect(onSaveMock).toHaveBeenCalled();
  });

  it("displays a toast message on form submission error", async () => {
    AppBackend.UpdateSSHSession.mockRejectedValue(new Error("Backend offline"));

    render(
      <SSHSessionForm
        session={mockSession}
        onClose={onCloseMock}
        onSave={onSaveMock}
        addToast={addToastMock}
      />,
    );

    const updateBtn = screen.getByRole("button", { name: "Update" });
    fireEvent.click(updateBtn);

    await waitFor(() => {
      expect(addToastMock).toHaveBeenCalledWith(
        "Error",
        expect.stringContaining(
          "Failed to save session: Error: Backend offline",
        ),
        "error",
      );
    });
  });
});
