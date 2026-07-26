import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { vi, describe, it, expect, beforeEach } from "vitest";
import ProxyTab from "./ProxyTab";
import * as AppBackend from "../../wailsjs/go/backend/App";

// Mock Wails backend functions
vi.mock("../../wailsjs/go/backend/App", () => ({
  GetProxyApps: vi.fn().mockResolvedValue([
    { name: "app1", port: 3000 },
    { name: "app2", port: 8080 },
  ]),
  SaveProxyPort: vi.fn().mockResolvedValue(null),
  OpenProxyTerminal: vi.fn(),
}));

describe("ProxyTab Component", () => {
  const mockProps = {
    addToast: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders and lists apps", async () => {
    render(<ProxyTab {...mockProps} />);

    expect(AppBackend.GetProxyApps).toHaveBeenCalled();

    await waitFor(() => {
      expect(screen.getByText("app1")).toBeInTheDocument();
      expect(screen.getByText("app2")).toBeInTheDocument();
    });
  });

  it("filters apps by search term", async () => {
    render(<ProxyTab {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("app1")).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText(/Search folders\.\.\./i);
    fireEvent.change(searchInput, { target: { value: "app1" } });

    expect(screen.getByText("app1")).toBeInTheDocument();
    expect(screen.queryByText("app2")).not.toBeInTheDocument();
  });

  it("handles port change and saving", async () => {
    render(<ProxyTab {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("app1")).toBeInTheDocument();
    });

    const input = screen.getAllByPlaceholderText(/e\.g\. 3000/i)[0];
    fireEvent.change(input, { target: { value: "4000" } });

    const saveBtn = screen.getAllByRole("button", { name: /Save/i })[0];
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(AppBackend.SaveProxyPort).toHaveBeenCalledWith("app1", 4000);
      expect(mockProps.addToast).toHaveBeenCalledWith(
        "Success",
        expect.any(String),
        "info",
      );
    });
  });

  it("opens terminal dropdown and clicks CMD", async () => {
    render(<ProxyTab {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("app1")).toBeInTheDocument();
    });

    const terminalBtn = screen.getAllByTitle("Terminal")[0];
    fireEvent.click(terminalBtn);

    expect(screen.getByText("CMD")).toBeInTheDocument();
    expect(screen.getByText("PowerShell")).toBeInTheDocument();

    fireEvent.click(screen.getByText("CMD"));
    expect(AppBackend.OpenProxyTerminal).toHaveBeenCalledWith("app1", "cmd");
  });

  it("opens terminal dropdown and clicks PowerShell", async () => {
    render(<ProxyTab {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("app1")).toBeInTheDocument();
    });

    const terminalBtn = screen.getAllByTitle("Terminal")[0];
    fireEvent.click(terminalBtn);

    expect(screen.getByText("PowerShell")).toBeInTheDocument();

    fireEvent.click(screen.getByText("PowerShell"));
    expect(AppBackend.OpenProxyTerminal).toHaveBeenCalledWith("app1", "powershell");
  });

  it("handles dropdown backdrop dismissal", async () => {
    render(<ProxyTab {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("app1")).toBeInTheDocument();
    });

    const terminalBtn = screen.getAllByTitle("Terminal")[0];
    fireEvent.click(terminalBtn); // Open dropdown

    const backdrop = document.querySelector(".fixed.inset-0.z-10");
    expect(backdrop).toBeInTheDocument();

    fireEvent.click(backdrop!); // Dismiss dropdown
    expect(screen.queryByText("CMD")).not.toBeInTheDocument();
  });

  it("handles save port failure case gracefully", async () => {
    vi.mocked(AppBackend.SaveProxyPort).mockRejectedValueOnce(new Error("Failed to save proxy port"));

    render(<ProxyTab {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText("app1")).toBeInTheDocument();
    });

    const saveBtn = screen.getAllByRole("button", { name: /Save/i })[0];
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockProps.addToast).toHaveBeenCalledWith("Error", "Error: Failed to save proxy port", "error");
    });
  });
});
