import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import CircularProgress from "./CircularProgress";
import React from "react";

describe("CircularProgress Component", () => {
  it("renders correctly with percentage", () => {
    render(<CircularProgress percentage={50} status="Downloading" />);
    expect(screen.getByText("50%")).toBeInTheDocument();
  });

  it("renders speed and downloaded amount", () => {
    render(
      <CircularProgress percentage={50} speed="1.5 MB/s" downloaded="10 MB" />,
    );
    expect(screen.getByText("1.5 MB/s")).toBeInTheDocument();
    expect(screen.getByText("10 MB")).toBeInTheDocument();
  });

  it("calls onCancel when clicked", () => {
    const onCancel = vi.fn();
    render(<CircularProgress percentage={50} onCancel={onCancel} />);

    const container = screen.getByText("50%").closest("div")
      .parentElement.parentElement;
    fireEvent.click(container);

    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it("shows loader when status is Streaming", () => {
    const { container } = render(
      <CircularProgress percentage={0} status="Streaming..." />,
    );
    // Lucide icons are SVGs. Let's look for the loader-circle class which is what Loader2 renders
    const loader = container.querySelector(".lucide-loader-circle");
    expect(loader).toBeInTheDocument();
    expect(loader).toHaveClass("animate-spin");
  });
});
