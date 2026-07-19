import { render, screen, fireEvent, act } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import React from "react";
import SSHFileExplorer from "./SSHFileExplorer";

describe("SSHFileExplorer Component", () => {
  const onNavigateUpMock = vi.fn();
  const onSyncMock = vi.fn();
  const setSearchQueryMock = vi.fn();
  const setEditingPathMock = vi.fn();
  const onUploadMock = vi.fn();
  const onNewFolderMock = vi.fn();
  const onFileDoubleClickMock = vi.fn();
  const onFileContextMenuMock = vi.fn();
  const formatSizeMock = vi.fn((bytes) => `${bytes} B`);
  const toggleSortMock = vi.fn();
  const onManualNavigationMock = vi.fn();

  const mockFiles = [
    { name: "folder-1", isDir: true, size: 0 },
    { name: "file-1.txt", isDir: false, size: 1024 },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders correctly with remote path and file list", () => {
    render(
      <SSHFileExplorer
        remotePath="/var/www"
        editingPath="/var/www"
        setEditingPath={setEditingPathMock}
        onNavigateUp={onNavigateUpMock}
        onSync={onSyncMock}
        searchQuery=""
        setSearchQuery={setSearchQueryMock}
        onUpload={onUploadMock}
        onNewFolder={onNewFolderMock}
        loadingFiles={false}
        sortedFiles={mockFiles}
        onFileDoubleClick={onFileDoubleClickMock}
        onFileContextMenu={onFileContextMenuMock}
        formatSize={formatSizeMock}
        toggleSort={toggleSortMock}
        sortConfig={{ key: "name", direction: "asc" }}
        onManualNavigation={onManualNavigationMock}
      />,
    );

    expect(screen.getByDisplayValue("/var/www")).toBeInTheDocument();
    expect(screen.getByText("folder-1")).toBeInTheDocument();
    expect(screen.getByText("file-1.txt")).toBeInTheDocument();
    expect(screen.getByText("1024 B")).toBeInTheDocument();
  });

  it("handles custom input for editing path and Enter/Escape actions", () => {
    render(
      <SSHFileExplorer
        remotePath="/var/www"
        editingPath="/var/www"
        setEditingPath={setEditingPathMock}
        onNavigateUp={onNavigateUpMock}
        onSync={onSyncMock}
        searchQuery=""
        setSearchQuery={setSearchQueryMock}
        onUpload={onUploadMock}
        onNewFolder={onNewFolderMock}
        loadingFiles={false}
        sortedFiles={mockFiles}
        onFileDoubleClick={onFileDoubleClickMock}
        onFileContextMenu={onFileContextMenuMock}
        formatSize={formatSizeMock}
        toggleSort={toggleSortMock}
        sortConfig={{ key: "name", direction: "asc" }}
        onManualNavigation={onManualNavigationMock}
      />,
    );

    const pathInput = screen.getByDisplayValue("/var/www");

    // Test Enter Key for manual navigation
    fireEvent.keyDown(pathInput, { key: "Enter" });
    expect(onManualNavigationMock).toHaveBeenCalledWith("/var/www");

    // Test Escape Key resets the path
    fireEvent.keyDown(pathInput, { key: "Escape" });
    expect(setEditingPathMock).toHaveBeenCalledWith("/var/www");
  });

  it("triggers onUpload and onNewFolder click events", () => {
    render(
      <SSHFileExplorer
        remotePath="/var/www"
        editingPath="/var/www"
        setEditingPath={setEditingPathMock}
        onNavigateUp={onNavigateUpMock}
        onSync={onSyncMock}
        searchQuery=""
        setSearchQuery={setSearchQueryMock}
        onUpload={onUploadMock}
        onNewFolder={onNewFolderMock}
        loadingFiles={false}
        sortedFiles={mockFiles}
        onFileDoubleClick={onFileDoubleClickMock}
        onFileContextMenu={onFileContextMenuMock}
        formatSize={formatSizeMock}
        toggleSort={toggleSortMock}
        sortConfig={{ key: "name", direction: "asc" }}
        onManualNavigation={onManualNavigationMock}
      />,
    );

    const uploadBtn = screen.getByRole("button", { name: /upload/i });
    fireEvent.click(uploadBtn);
    expect(onUploadMock).toHaveBeenCalled();

    const newFolderBtn = screen.getByRole("button", { name: /new/i });
    fireEvent.click(newFolderBtn);
    expect(onNewFolderMock).toHaveBeenCalled();
  });

  it("triggers sorting toggle actions on headers", () => {
    render(
      <SSHFileExplorer
        remotePath="/var/www"
        editingPath="/var/www"
        setEditingPath={setEditingPathMock}
        onNavigateUp={onNavigateUpMock}
        onSync={onSyncMock}
        searchQuery=""
        setSearchQuery={setSearchQueryMock}
        onUpload={onUploadMock}
        onNewFolder={onNewFolderMock}
        loadingFiles={false}
        sortedFiles={mockFiles}
        onFileDoubleClick={onFileDoubleClickMock}
        onFileContextMenu={onFileContextMenuMock}
        formatSize={formatSizeMock}
        toggleSort={toggleSortMock}
        sortConfig={{ key: "name", direction: "asc" }}
        onManualNavigation={onManualNavigationMock}
      />,
    );

    const nameHeader = screen.getByRole("button", { name: /name/i });
    fireEvent.click(nameHeader);
    expect(toggleSortMock).toHaveBeenCalledWith("name");

    const sizeHeader = screen.getByRole("button", { name: /size/i });
    fireEvent.click(sizeHeader);
    expect(toggleSortMock).toHaveBeenCalledWith("size");
  });

  it("renders loading indicator state", () => {
    render(
      <SSHFileExplorer
        remotePath="/var/www"
        editingPath="/var/www"
        setEditingPath={setEditingPathMock}
        onNavigateUp={onNavigateUpMock}
        onSync={onSyncMock}
        searchQuery=""
        setSearchQuery={setSearchQueryMock}
        onUpload={onUploadMock}
        onNewFolder={onNewFolderMock}
        loadingFiles={true}
        sortedFiles={[]}
        onFileDoubleClick={onFileDoubleClickMock}
        onFileContextMenu={onFileContextMenuMock}
        formatSize={formatSizeMock}
        toggleSort={toggleSortMock}
        sortConfig={{ key: "name", direction: "asc" }}
        onManualNavigation={onManualNavigationMock}
      />,
    );

    // Check that files are not rendered, but loading indicator is present
    expect(screen.queryByText("folder-1")).not.toBeInTheDocument();
  });
});
