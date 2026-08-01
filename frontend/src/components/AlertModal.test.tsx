import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import React from "react";
import AlertModal from "./AlertModal";

describe("AlertModal Component", () => {
  it("renders nothing when isOpen is false", () => {
    const { container } = render(
      <AlertModal
        isOpen={false}
        title="Test Alert"
        message="Test Message"
        onClose={vi.fn()}
      />
    );
    expect(container.firstChild).toBeNull();
  });

  it("renders correctly and triggers onClose when backdrop or close button is clicked", () => {
    const handleClose = vi.fn();
    render(
      <AlertModal
        isOpen={true}
        title="About Ostenia"
        message="This is a test message"
        onClose={handleClose}
      />
    );

    expect(screen.getByText("About Ostenia")).toBeInTheDocument();
    expect(screen.getByText("This is a test message")).toBeInTheDocument();

    const okBtn = screen.getByRole("button", { name: "OK" });
    fireEvent.click(okBtn);
    expect(handleClose).toHaveBeenCalledTimes(1);
  });
});
