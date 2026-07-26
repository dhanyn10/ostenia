import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import AddCustomVersionModal from "./AddCustomVersionModal";
import React from "react";

// Mock AppBackend functions
vi.mock("../../wailsjs/go/backend/App", () => ({
  ProcessCustomVersion: vi.fn().mockResolvedValue(null),
  ProcessCustomVersionBytes: vi.fn().mockResolvedValue(null),
  OpenFileDialog: vi.fn().mockResolvedValue("C:\\test\\file.zip"),
  OpenDirectoryDialog: vi.fn().mockResolvedValue("C:\\test\\folder"),
}));

import * as AppBackend from "../../wailsjs/go/backend/App";

describe("AddCustomVersionModal Component", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("does not render when isOpen is false", () => {
    const { container } = render(
      <AddCustomVersionModal
        isOpen={false}
        onClose={() => {}}
        serviceName="PHP"
        onSuccess={() => {}}
        appsLocation="C:\\Ostenia"
      />,
    );
    expect(container.firstChild).toBeNull();
  });

  it("renders correctly when open", () => {
    render(
      <AddCustomVersionModal
        isOpen={true}
        onClose={() => {}}
        serviceName="PHP"
        onSuccess={() => {}}
        appsLocation="C:\\Ostenia"
      />,
    );
    expect(screen.getByText("Add Custom PHP Version")).toBeInTheDocument();
    expect(screen.getByText("Select ZIP File")).toBeInTheDocument();
    expect(screen.getByText("Select Folder")).toBeInTheDocument();
  });

  it("calls onClose when Close button is clicked", () => {
    const onClose = vi.fn();
    render(
      <AddCustomVersionModal
        isOpen={true}
        onClose={onClose}
        serviceName="PHP"
        onSuccess={() => {}}
        appsLocation="C:\\Ostenia"
      />,
    );
    fireEvent.click(screen.getByText("Close"));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("handles manual ZIP file selection successfully", async () => {
    const onSuccess = vi.fn();
    render(
      <AddCustomVersionModal
        isOpen={true}
        onClose={() => {}}
        serviceName="PHP"
        onSuccess={onSuccess}
        appsLocation="C:\\Ostenia"
      />,
    );

    fireEvent.click(screen.getByText("Select ZIP File"));

    await waitFor(() => {
      expect(AppBackend.OpenFileDialog).toHaveBeenCalledTimes(1);
      expect(AppBackend.ProcessCustomVersion).toHaveBeenCalledWith("PHP", "C:\\test\\file.zip");
      expect(onSuccess).toHaveBeenCalledTimes(1);
    });
  });

  it("handles manual folder selection successfully", async () => {
    const onSuccess = vi.fn();
    render(
      <AddCustomVersionModal
        isOpen={true}
        onClose={() => {}}
        serviceName="PHP"
        onSuccess={onSuccess}
        appsLocation="C:\\Ostenia"
      />,
    );

    fireEvent.click(screen.getByText("Select Folder"));

    await waitFor(() => {
      expect(AppBackend.OpenDirectoryDialog).toHaveBeenCalledTimes(1);
      expect(AppBackend.ProcessCustomVersion).toHaveBeenCalledWith("PHP", "C:\\test\\folder");
      expect(onSuccess).toHaveBeenCalledTimes(1);
    });
  });

  it("handles drag-over and drag-leave events correctly", () => {
    render(
      <AddCustomVersionModal
        isOpen={true}
        onClose={() => {}}
        serviceName="PHP"
        onSuccess={() => {}}
        appsLocation="C:\\Ostenia"
      />,
    );

    const dropzone = screen.getByText("Drag & Drop ZIP / Folder here").closest(".wails-drop-target")!;

    // Drag enter/over
    fireEvent.dragEnter(dropzone);
    expect(dropzone).toHaveClass("border-blue-500");

    // Drag leave
    fireEvent.dragLeave(dropzone);
    expect(dropzone).not.toHaveClass("border-blue-500");
  });

  it("shows error when dropping a non-ZIP file", async () => {
    render(
      <AddCustomVersionModal
        isOpen={true}
        onClose={() => {}}
        serviceName="PHP"
        onSuccess={() => {}}
        appsLocation="C:\\Ostenia"
      />,
    );

    const dropzone = screen.getByText("Drag & Drop ZIP / Folder here").closest(".wails-drop-target")!;

    const file = new File(["test-content"], "test-file.txt", { type: "text/plain" });
    fireEvent.drop(dropzone, {
      dataTransfer: {
        files: [file],
      },
    });

    await waitFor(() => {
      expect(screen.getByText("For custom folder addition, please use the 'Select Folder' button to select and copy the folder directly.")).toBeInTheDocument();
    });
  });

  it("handles dropping a valid ZIP file and processes bytes", async () => {
    const onSuccess = vi.fn();
    render(
      <AddCustomVersionModal
        isOpen={true}
        onClose={() => {}}
        serviceName="PHP"
        onSuccess={onSuccess}
        appsLocation="C:\\Ostenia"
      />,
    );

    const dropzone = screen.getByText("Drag & Drop ZIP / Folder here").closest(".wails-drop-target")!;

    const file = new File(["zip-content"], "test-file.zip", { type: "application/zip" });
    file.arrayBuffer = vi.fn().mockResolvedValue(new ArrayBuffer(11));

    fireEvent.drop(dropzone, {
      dataTransfer: {
        files: [file],
      },
    });

    await waitFor(() => {
      expect(AppBackend.ProcessCustomVersionBytes).toHaveBeenCalled();
      expect(onSuccess).toHaveBeenCalledTimes(1);
    });
  });

  it("displays error message if manual dialog selection throws an error", async () => {
    vi.spyOn(AppBackend, "OpenFileDialog").mockRejectedValueOnce(new Error("User cancelled"));

    render(
      <AddCustomVersionModal
        isOpen={true}
        onClose={() => {}}
        serviceName="PHP"
        onSuccess={() => {}}
        appsLocation="C:\\Ostenia"
      />,
    );

    fireEvent.click(screen.getByText("Select ZIP File"));

    await waitFor(() => {
      expect(screen.getByText("User cancelled")).toBeInTheDocument();
    });
  });
});
