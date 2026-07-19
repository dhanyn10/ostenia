import React from "react";
import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import AddPluginAction from "./AddPluginAction";

describe("AddPluginAction Component", () => {
  const mockPrerequisites = [
    { name: "PHP" },
    { name: "MySQL" },
    { name: "Nginx" },
  ];

  const mockServices = [
    { name: "PHP" }, // PHP is already pinned, so only MySQL and Nginx should be shown when adding
  ];

  const defaultProps = {
    isAddingPlugin: false,
    setIsAddingPlugin: vi.fn(),
    prerequisites: mockPrerequisites,
    services: mockServices,
    handleAddToHome: vi.fn(),
    renderIcon: (name: string) => <div data-testid={`icon-${name}`} />,
  };

  it("renders correctly in closed state", () => {
    render(<AddPluginAction {...defaultProps} />);

    expect(screen.getByText("Add Plugin to Home")).toBeInTheDocument();
    expect(screen.queryByText("Close Menu")).not.toBeInTheDocument();
    expect(screen.queryByText("MySQL")).not.toBeInTheDocument();
    expect(screen.queryByText("Nginx")).not.toBeInTheDocument();
  });

  it("renders correctly in open state with available plugins", () => {
    render(<AddPluginAction {...defaultProps} isAddingPlugin={true} />);

    expect(screen.getByText("Close Menu")).toBeInTheDocument();
    expect(screen.queryByText("Add Plugin to Home")).not.toBeInTheDocument();

    // PHP is already pinned, so it should NOT be in the options
    expect(screen.queryByText("PHP")).not.toBeInTheDocument();

    // MySQL and Nginx are not pinned, so they should be shown
    expect(screen.getByText("MySQL")).toBeInTheDocument();
    expect(screen.getByText("Nginx")).toBeInTheDocument();

    // Icons should be rendered correctly
    expect(screen.getByTestId("icon-MySQL")).toBeInTheDocument();
    expect(screen.getByTestId("icon-Nginx")).toBeInTheDocument();
  });

  it("calls setIsAddingPlugin when clicking the main toggle button", () => {
    const setIsAddingPlugin = vi.fn();
    render(
      <AddPluginAction
        {...defaultProps}
        setIsAddingPlugin={setIsAddingPlugin}
      />,
    );

    const button = screen.getByRole("button");
    fireEvent.click(button);

    expect(setIsAddingPlugin).toHaveBeenCalledWith(true);
  });

  it("calls handleAddToHome when clicking on an available plugin option", () => {
    const handleAddToHome = vi.fn();
    render(
      <AddPluginAction
        {...defaultProps}
        isAddingPlugin={true}
        handleAddToHome={handleAddToHome}
      />,
    );

    const mysqlButton = screen.getByRole("button", { name: /MySQL/i });
    fireEvent.click(mysqlButton);

    expect(handleAddToHome).toHaveBeenCalledWith({ name: "MySQL" });
  });

  it("shows a message when all plugins are already pinned", () => {
    const allServices = [{ name: "PHP" }, { name: "MySQL" }, { name: "Nginx" }];

    render(
      <AddPluginAction
        {...defaultProps}
        isAddingPlugin={true}
        services={allServices}
      />,
    );

    expect(
      screen.getByText("All plugins are already pinned"),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /MySQL/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /PHP/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /Nginx/i }),
    ).not.toBeInTheDocument();
  });
});
