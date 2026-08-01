import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import React from "react";
import PromptModal from "./PromptModal";

describe("PromptModal Component", () => {
  it("renders nothing when isOpen is false", () => {
    const { container } = render(
      <PromptModal
        isOpen={false}
        title="Test Prompt"
        message="Enter value:"
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    );
    expect(container.firstChild).toBeNull();
  });

  it("handles input submission and cancellation", async () => {
    const handleConfirm = vi.fn();
    const handleCancel = vi.fn();

    render(
      <PromptModal
        isOpen={true}
        title="Rename"
        message="Enter new file name:"
        defaultValue="old-name.txt"
        placeholder="New name"
        onConfirm={handleConfirm}
        onCancel={handleCancel}
      />
    );

    expect(screen.getByText("Rename")).toBeInTheDocument();
    expect(screen.getByText("Enter new file name:")).toBeInTheDocument();

    const input = screen.getByPlaceholderText("New name") as HTMLInputElement;
    expect(input.value).toBe("old-name.txt");

    fireEvent.change(input, { target: { value: "new-name.txt" } });
    expect(input.value).toBe("new-name.txt");

    const submitBtn = screen.getByRole("button", { name: "Save" });
    fireEvent.click(submitBtn);

    expect(handleConfirm).toHaveBeenCalledWith("new-name.txt");

    // Test cancellation
    const cancelBtn = screen.getByRole("button", { name: "Cancel" });
    fireEvent.click(cancelBtn);
    expect(handleCancel).toHaveBeenCalledTimes(1);
  });
});
