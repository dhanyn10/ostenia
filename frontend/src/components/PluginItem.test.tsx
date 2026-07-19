import { render, screen, fireEvent } from "@testing-library/react";
import { vi, describe, it, expect, beforeEach } from "vitest";
import PluginItem from "./PluginItem";

describe("PluginItem Component", () => {
  const mockProps = {
    task: {
      name: "PHP",
      version: "8.2",
      isInstalled: true,
      installedVers: ["8.1", "8.2"],
      versions: ["8.1", "8.2", "8.3"],
      info: "Highly recommended",
      modules: [
        {
          name: "xdebug",
          isInstalled: false,
          status: "Not Installed",
          version: "3.2",
        },
      ],
    },
    progress: {},
    isDropdownOpen: false,
    onDropdownToggle: vi.fn(),
    selectedVersion: "8.2",
    onVersionChange: vi.fn(),
    onDeleteVersion: vi.fn(),
    onInstall: vi.fn(),
    onCancel: vi.fn(),
    onOpenFolder: vi.fn(),
    renderIcon: vi.fn(() => <span data-testid="icon" />),
    onInstallModule: vi.fn(),
    onUninstallModule: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders plugin information correctly", () => {
    render(<PluginItem {...mockProps} />);

    expect(screen.getByText("PHP")).toBeInTheDocument();
    expect(screen.getByText("Highly recommended")).toBeInTheDocument();
    expect(screen.getByTestId("icon")).toBeInTheDocument();
  });

  it("shows download button when version not installed", () => {
    const props = {
      ...mockProps,
      selectedVersion: "8.3",
    };
    render(<PluginItem {...props} />);

    expect(
      screen.getByRole("button", { name: /Download/i }),
    ).toBeInTheDocument();
  });

  it("shows ready status when version is installed", () => {
    render(<PluginItem {...mockProps} />);

    expect(screen.getByRole("button", { name: /Ready/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Ready/i })).toBeDisabled();
  });

  it("handles version dropdown toggle", () => {
    render(<PluginItem {...mockProps} />);

    const versionDisplay = screen.getByText("v8.2");
    fireEvent.click(versionDisplay);

    expect(mockProps.onDropdownToggle).toHaveBeenCalled();
  });

  it("handles version deletion", () => {
    render(<PluginItem {...mockProps} />);

    const deleteButton = screen.getByTitle("Delete v8.1");
    fireEvent.click(deleteButton);

    expect(mockProps.onDeleteVersion).toHaveBeenCalledWith("PHP", "8.1");
  });

  it("expands modules and handles module installation", () => {
    render(<PluginItem {...mockProps} />);

    const buttons = screen.getAllByRole("button");
    const chevronButton = buttons.find((b) =>
      b.querySelector("svg.lucide-chevron-down"),
    );
    fireEvent.click(chevronButton!);

    expect(screen.getByText("Available Modules")).toBeInTheDocument();
    expect(screen.getByText("xdebug")).toBeInTheDocument();

    const installModuleBtn = screen.getByTitle("Install xdebug");
    fireEvent.click(installModuleBtn);

    expect(mockProps.onInstallModule).toHaveBeenCalledWith("PHP", "xdebug");
  });

  it("shows circular progress during active download", () => {
    const props = {
      ...mockProps,
      progress: {
        PHP: {
          percentage: 45,
          status: "Downloading",
          speed: "2MB/s",
          downloaded: "10MB",
        },
      },
    };
    render(<PluginItem {...props} />);

    expect(screen.getByText("45%")).toBeInTheDocument();
  });
});
