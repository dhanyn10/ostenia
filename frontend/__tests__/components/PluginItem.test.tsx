import { render, screen, fireEvent } from "@testing-library/react";
import { vi, describe, it, expect, beforeEach } from "vitest";
import PluginItem from "@/components/PluginItem";

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

    fireEvent.keyDown(deleteButton, { key: "Enter" });
    expect(mockProps.onDeleteVersion).toHaveBeenCalledTimes(2);
  });

  it("renders disabled module button when main task is not installed", () => {
    const uninstalledTaskProps = {
      ...mockProps,
      task: {
        ...mockProps.task,
        isInstalled: false,
      },
    };
    render(<PluginItem {...uninstalledTaskProps} />);

    const card = document.querySelector('[data-card-expand]') as HTMLElement;
    fireEvent.click(card);

    expect(screen.getByTitle("Install PHP first")).toBeInTheDocument();
  });

  it("handles module uninstallation when module is already installed", () => {
    const installedModuleProps = {
      ...mockProps,
      task: {
        ...mockProps.task,
        modules: [
          {
            name: "xdebug",
            isInstalled: true,
            status: "Installed",
            version: "3.2",
          },
        ],
      },
    };
    render(<PluginItem {...installedModuleProps} />);

    const card = document.querySelector('[data-card-expand]') as HTMLElement;
    fireEvent.click(card);

    expect(screen.getByTitle("Installed")).toBeInTheDocument();
    const uninstallBtn = screen.getByTitle("Uninstall xdebug");
    fireEvent.click(uninstallBtn);
    expect(mockProps.onUninstallModule).toHaveBeenCalledWith("PHP", "xdebug");
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

  it("expands modules when the card element is clicked", () => {
    const { container } = render(<PluginItem {...mockProps} />);

    expect(screen.queryByText("Available Modules")).not.toBeInTheDocument();

    const card = container.querySelector('[data-card-expand]') as HTMLElement;
    fireEvent.click(card);

    expect(screen.getByText("Available Modules")).toBeInTheDocument();
    expect(screen.getByText("xdebug")).toBeInTheDocument();

    // Clicking card again collapses it
    fireEvent.click(card);
    expect(screen.queryByText("Available Modules")).not.toBeInTheDocument();
  });

  it("supports keyboard expansion with Enter and Space on the card", () => {
    const { container } = render(<PluginItem {...mockProps} />);
    const card = container.querySelector('[data-card-expand]') as HTMLElement;

    expect(screen.queryByText("Available Modules")).not.toBeInTheDocument();

    fireEvent.click(card);
    expect(screen.getByText("Available Modules")).toBeInTheDocument();

    fireEvent.click(card);
    expect(screen.queryByText("Available Modules")).not.toBeInTheDocument();
  });

  it("does not expand on click or key press when plugin has no modules", () => {
    const propsNoModules = {
      ...mockProps,
      task: {
        ...mockProps.task,
        modules: [],
      },
    };
    const { container } = render(<PluginItem {...propsNoModules} />);
    expect(container.querySelector('[data-card-expand]')).toBeNull();
    expect(screen.queryByText("Available Modules")).not.toBeInTheDocument();
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

  it("auto-expands modules when clicking anywhere on the card (title, info, card root)", () => {
    const { container } = render(<PluginItem {...mockProps} />);

    expect(screen.queryByText("Available Modules")).not.toBeInTheDocument();

    // 1. Click on plugin title "PHP" -> expands
    fireEvent.click(screen.getByText("PHP"));
    expect(screen.getByText("Available Modules")).toBeInTheDocument();

    // 2. Click on card root -> collapses
    const cardRoot = container.querySelector('[data-card-root="true"]') as HTMLElement;
    fireEvent.click(cardRoot);
    expect(screen.queryByText("Available Modules")).not.toBeInTheDocument();

    // 3. Click on info badge "Highly recommended" -> expands
    fireEvent.click(screen.getByText("Highly recommended"));
    expect(screen.getByText("Available Modules")).toBeInTheDocument();

    // 4. Clicking inside module list does NOT collapse the card
    fireEvent.click(screen.getByText("Available Modules"));
    expect(screen.getByText("Available Modules")).toBeInTheDocument();
  });

  it("does not toggle card expansion when interactive controls like version dropdown or delete version are clicked", () => {
    render(<PluginItem {...mockProps} />);

    // Click version dropdown
    fireEvent.click(screen.getByText("v8.2"));
    expect(mockProps.onDropdownToggle).toHaveBeenCalled();
    expect(screen.queryByText("Available Modules")).not.toBeInTheDocument();

    // Click delete version button
    const deleteBtn = screen.getByTitle("Delete v8.1");
    fireEvent.click(deleteBtn);
    expect(mockProps.onDeleteVersion).toHaveBeenCalledWith("PHP", "8.1");
    expect(screen.queryByText("Available Modules")).not.toBeInTheDocument();
  });
});
