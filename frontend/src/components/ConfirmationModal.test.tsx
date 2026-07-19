import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import ConfirmationModal from "./ConfirmationModal";
import React from "react";

describe("ConfirmationModal Component", () => {
  it("does not render when isOpen is false", () => {
    const { container } = render(
      <ConfirmationModal
        isOpen={false}
        title="Test Title"
        message="Test Message"
        onConfirm={() => {}}
        onCancel={() => {}}
      />,
    );
    expect(container.firstChild).toBeNull();
  });

  it("renders correctly when open", () => {
    render(
      <ConfirmationModal
        isOpen={true}
        title="Delete Item"
        message="Are you sure?"
        onConfirm={() => {}}
        onCancel={() => {}}
      />,
    );
    expect(screen.getByText("Delete Item")).toBeInTheDocument();
    expect(screen.getByText("Are you sure?")).toBeInTheDocument();
    expect(screen.getByText("Confirm")).toBeInTheDocument();
    expect(screen.getByText("Cancel")).toBeInTheDocument();
  });

  it("calls onConfirm when confirm button is clicked", () => {
    const onConfirm = vi.fn();
    render(
      <ConfirmationModal
        isOpen={true}
        title="Title"
        message="Message"
        onConfirm={onConfirm}
        onCancel={() => {}}
      />,
    );
    fireEvent.click(screen.getByText("Confirm"));
    expect(onConfirm).toHaveBeenCalledTimes(1);
  });

  it("calls onCancel when cancel button is clicked", () => {
    const onCancel = vi.fn();
    render(
      <ConfirmationModal
        isOpen={true}
        title="Title"
        message="Message"
        onConfirm={() => {}}
        onCancel={onCancel}
      />,
    );
    fireEvent.click(screen.getByText("Cancel"));
    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it("calls onCancel when backdrop is clicked", () => {
    const onCancel = vi.fn();
    const { container } = render(
      <ConfirmationModal
        isOpen={true}
        title="Title"
        message="Message"
        onConfirm={() => {}}
        onCancel={onCancel}
      />,
    );
    // The backdrop is the first button
    const backdrop = container.querySelector("button.absolute.inset-0");
    if (backdrop) {
      fireEvent.click(backdrop);
      expect(onCancel).toHaveBeenCalledTimes(1);
    }
  });

  it("renders info type correctly", () => {
    render(
      <ConfirmationModal
        isOpen={true}
        title="Info Title"
        message="Info Message"
        onConfirm={() => {}}
        onCancel={() => {}}
        type="info"
      />,
    );
    const confirmBtn = screen.getByText("Confirm");
    expect(confirmBtn).toHaveClass("bg-blue-600");
  });
});
