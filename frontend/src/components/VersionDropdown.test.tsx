import React from "react";
import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import VersionDropdown from "./VersionDropdown";

describe("VersionDropdown Component", () => {
  it("renders correctly with current version and closed state", () => {
    const onChange = vi.fn();
    const onToggle = vi.fn();
    render(
      <VersionDropdown
        current="8.3"
        options={["8.1", "8.2", "8.3"]}
        onChange={onChange}
        isOpen={false}
        onToggle={onToggle}
      />,
    );

    expect(screen.getByText("v8.3")).toBeInTheDocument();
    expect(screen.queryByText("v8.1")).not.toBeInTheDocument();
    expect(screen.queryByText("v8.2")).not.toBeInTheDocument();
  });

  it("calls onToggle when clicking the main button", () => {
    const onToggle = vi.fn();
    render(
      <VersionDropdown
        current="8.3"
        options={["8.1", "8.2", "8.3"]}
        onChange={vi.fn()}
        isOpen={false}
        onToggle={onToggle}
      />,
    );

    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(onToggle).toHaveBeenCalledTimes(1);
  });

  it("renders all options and custom option when open", () => {
    render(
      <VersionDropdown
        current="8.3"
        options={["8.1", "8.2", "8.3"]}
        onChange={vi.fn()}
        isOpen={true}
        onToggle={vi.fn()}
        allowCustom={true}
        onCustomClick={vi.fn()}
      />,
    );

    expect(screen.getByText("v8.1")).toBeInTheDocument();
    expect(screen.getByText("v8.2")).toBeInTheDocument();

    // There should be two elements with 'v8.3' (the button and the dropdown option)
    const elements = screen.getAllByText("v8.3");
    expect(elements).toHaveLength(2);

    expect(screen.getByText("Add Custom...")).toBeInTheDocument();
  });

  it("calls onChange and onToggle when an option is clicked", () => {
    const onChange = vi.fn();
    const onToggle = vi.fn();
    render(
      <VersionDropdown
        current="8.3"
        options={["8.1", "8.2", "8.3"]}
        onChange={onChange}
        isOpen={true}
        onToggle={onToggle}
      />,
    );

    const optionBtn = screen.getByRole("button", { name: "v8.1" });
    fireEvent.click(optionBtn);

    expect(onChange).toHaveBeenCalledWith("8.1");
    expect(onToggle).toHaveBeenCalledTimes(1);
  });

  it("calls onCustomClick and onToggle when custom button is clicked", () => {
    const onCustomClick = vi.fn();
    const onToggle = vi.fn();
    render(
      <VersionDropdown
        current="8.3"
        options={["8.1", "8.2", "8.3"]}
        onChange={vi.fn()}
        isOpen={true}
        onToggle={onToggle}
        allowCustom={true}
        onCustomClick={onCustomClick}
      />,
    );

    const customBtn = screen.getByRole("button", { name: "Add Custom..." });
    fireEvent.click(customBtn);

    expect(onCustomClick).toHaveBeenCalledTimes(1);
    expect(onToggle).toHaveBeenCalledTimes(1);
  });
});
