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
});
