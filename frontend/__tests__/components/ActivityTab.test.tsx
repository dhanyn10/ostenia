import { render, screen, fireEvent, waitFor, act } from "@testing-library/react";
import { vi, describe, it, expect, beforeEach } from "vitest";
import ActivityTab from "@/components/ActivityTab";
import React from "react";
import * as AppBackend from "@/wailsjs/go/backend/App";

// Mock Wails backend functions
vi.mock("@/wailsjs/go/backend/App", () => ({
  GetPHPExtensions: vi
    .fn()
    .mockResolvedValue([{ name: "mysqli", enabled: true }]),
  TogglePHPExtension: vi.fn().mockResolvedValue(null),
  SwitchServiceVersion: vi.fn().mockResolvedValue(null),
  OpenServiceTerminal: vi.fn().mockResolvedValue(null),
}));

describe("ActivityTab Component", () => {
  const defaultProps = {
    serverRoot: "/server/root",
    appsLocation: "/apps/location",
    handleBrowseAppsLocation: vi.fn(),
    handleBrowseServerRoot: vi.fn(),
    isAddingPlugin: false,
    setIsAddingPlugin: vi.fn(),
    prerequisites: [
      {
        name: "PHP",
        version: "8.2",
        installedVers: ["8.2.0"],
        status: "Ready",
      },
    ],
    services: [{ name: "PHP", status: "Running", port: 9000 }],
    handleAddToHome: vi.fn(),
    renderIcon: () => <span>Icon</span>,
    handleToggleService: vi.fn(),
    handleRemoveFromHome: vi.fn(),
    setActiveTab: vi.fn(),
    handleOpenPluginFolder: vi.fn(),
    handleOpenServerRootFolder: vi.fn(),
    handleOpenAppsLocationFolder: vi.fn(),
    apacheHttps: false,
    nginxHttps: false,
    handleToggleHttps: vi.fn(),
    isLoading: false,
    transitioningServices: new Set(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders correctly with services", () => {
    render(<ActivityTab {...defaultProps} />);
    expect(screen.getByText("Apps Location")).toBeInTheDocument();
    expect(screen.getByText("Server Root Directory")).toBeInTheDocument();
    expect(screen.getByText("PHP")).toBeInTheDocument();
  });

  it("calls browse handlers when buttons are clicked", () => {
    render(<ActivityTab {...defaultProps} />);

    const browseButtons = screen.getAllByTitle("Browse Directory");
    fireEvent.click(browseButtons[0]);
    expect(defaultProps.handleBrowseAppsLocation).toHaveBeenCalled();

    fireEvent.click(browseButtons[1]);
    expect(defaultProps.handleBrowseServerRoot).toHaveBeenCalled();
  });

  it("calls open folder handlers", () => {
    render(<ActivityTab {...defaultProps} />);

    const openButtons = screen.getAllByTitle("Open in Explorer");
    fireEvent.click(openButtons[0]);
    expect(defaultProps.handleOpenAppsLocationFolder).toHaveBeenCalled();

    fireEvent.click(openButtons[1]);
    expect(defaultProps.handleOpenServerRootFolder).toHaveBeenCalled();
  });

  it("fetches PHP extensions when PHP accordion is expanded", async () => {
    render(<ActivityTab {...defaultProps} />);

    const phpItem = screen.getByText("PHP");
    fireEvent.click(phpItem);

    await waitFor(() => {
      expect(AppBackend.GetPHPExtensions).toHaveBeenCalled();
    });
  });

  it("shows loading state", () => {
    render(<ActivityTab {...defaultProps} isLoading={true} />);
    expect(screen.getByText(/Scanning Plugins/i)).toBeInTheDocument();
  });

  it("shows empty state when no services", () => {
    render(<ActivityTab {...defaultProps} services={[]} />);
    expect(screen.getByText("No services active")).toBeInTheDocument();
  });

  it("handles version switching, terminal opening, and extension toggling", async () => {
    const propsWithMultipleVersions = {
      ...defaultProps,
      prerequisites: [
        {
          name: "PHP",
          version: "8.2",
          installedVers: ["8.2.0", "8.3.0"],
          status: "Ready",
        },
      ],
    };

    render(<ActivityTab {...propsWithMultipleVersions} />);

    // Expand PHP accordion
    fireEvent.click(screen.getByText("PHP"));

    // 1. Version switching
    await waitFor(() => {
      expect(screen.getByText("8.3.0")).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText("8.3.0"));
    expect(AppBackend.SwitchServiceVersion).toHaveBeenCalledWith("PHP", "8.3.0");

    // 2. Open local terminal
    const terminalBtn = screen.getByTitle("Terminal");
    fireEvent.click(terminalBtn);
    expect(screen.getByText("CMD")).toBeInTheDocument();
    fireEvent.click(screen.getByText("CMD"));
    expect(AppBackend.OpenServiceTerminal).toHaveBeenCalledWith("PHP", "cmd");

    // 3. Open Extensions modal and toggle
    fireEvent.click(screen.getByText("Extensions"));
    await waitFor(() => {
      expect(screen.getByText("mysqli")).toBeInTheDocument();
    });

    // Find the toggle button inside ExtensionModal
    const extToggle = screen.getByRole("checkbox", { hidden: true }) || screen.getByText("mysqli");
    await act(async () => {
      fireEvent.click(extToggle);
    });
    expect(AppBackend.TogglePHPExtension).toHaveBeenCalled();
  });
});
