import React from "react";
import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import PluginsTab from "./PluginsTab";

// Mock the child component PluginItem to isolate PluginsTab
vi.mock("./PluginItem", () => ({
  default: ({ task, onOpenFolder }: any) => (
    <div data-testid={`plugin-item-${task.name}`}>
      <span>Plugin: {task.name}</span>
      <button
        data-testid={`btn-folder-${task.name}`}
        onClick={() => onOpenFolder(task.name)}
      >
        Open Folder
      </button>
    </div>
  ),
}));

// Mock the AppBackend API imported inside PluginsTab
vi.mock("../../wailsjs/go/backend/App", () => ({
  OpenPluginFolder: vi.fn().mockResolvedValue(null),
  OpenPluginsConfigFolder: vi.fn().mockResolvedValue(null),
}));

describe("PluginsTab Component", () => {
  const mockPrerequisites = [
    { name: "PHP", version: "8.3" },
    { name: "MySQL", version: "8.0" },
  ];

  const mockOnAddCustomVersion = vi.fn();
  const mockOnReloadPlugins = vi.fn();
  const mockOnOpenPluginsFolder = vi.fn();

  const defaultProps = {
    prerequisites: mockPrerequisites,
    downloadProgress: {},
    openDropdown: null,
    setOpenDropdown: vi.fn(),
    selectedVersions: { PHP: "8.3", MySQL: "8.0" },
    setSelectedVersions: vi.fn(),
    handleDeleteVersion: vi.fn(),
    handleInstallSingle: vi.fn(),
    handleCancel: vi.fn(),
    renderIcon: vi.fn(),
    handleInstallModule: vi.fn(),
    handleUninstallModule: vi.fn(),
    onAddCustomVersion: mockOnAddCustomVersion,
    onReloadPlugins: mockOnReloadPlugins,
    onOpenPluginsFolder: mockOnOpenPluginsFolder,
  };

  it("renders a list of plugin items based on prerequisites", () => {
    render(<PluginsTab {...defaultProps} />);

    expect(screen.getByTestId("plugin-item-PHP")).toBeInTheDocument();
    expect(screen.getByTestId("plugin-item-MySQL")).toBeInTheDocument();
    expect(screen.getByText("Plugin: PHP")).toBeInTheDocument();
    expect(screen.getByText("Plugin: MySQL")).toBeInTheDocument();
  });

  it("handles null or empty tasks gracefully", () => {
    const tasksWithNull = [
      { name: "PHP", version: "8.3" },
      null,
      { name: "MySQL", version: "8.0" },
    ] as any;

    render(<PluginsTab {...defaultProps} prerequisites={tasksWithNull} />);

    expect(screen.getByTestId("plugin-item-PHP")).toBeInTheDocument();
    expect(screen.getByTestId("plugin-item-MySQL")).toBeInTheDocument();
    expect(screen.queryByTestId("plugin-item-null")).not.toBeInTheDocument();
  });

  it("calls onAddCustomVersion when clicking Open Folder in PluginItem", () => {
    render(<PluginsTab {...defaultProps} />);

    const folderBtn = screen.getByTestId("btn-folder-PHP");
    fireEvent.click(folderBtn);

    expect(mockOnAddCustomVersion).toHaveBeenCalledWith("PHP");
  });

  it("calls onReloadPlugins and onOpenPluginsFolder when toolbar buttons are clicked", () => {
    render(<PluginsTab {...defaultProps} />);

    const reloadBtn = screen.getByRole("button", { name: /Reload Plugins/i });
    fireEvent.click(reloadBtn);
    expect(mockOnReloadPlugins).toHaveBeenCalled();

    const openPluginsFolderBtn = screen.getByRole("button", { name: /Open Plugins Folder/i });
    fireEvent.click(openPluginsFolderBtn);
    expect(mockOnOpenPluginsFolder).toHaveBeenCalled();
  });
});
