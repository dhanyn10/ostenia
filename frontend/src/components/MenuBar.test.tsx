import { render, screen, fireEvent } from "@testing-library/react";
import { vi, describe, it, expect, beforeEach } from "vitest";
import MenuBar from "./MenuBar";
import * as AppBackend from "../../wailsjs/go/backend/App";

// Mock Wails backend functions
vi.mock("../../wailsjs/go/backend/App", () => ({
  Minimize: vi.fn(),
  Maximize: vi.fn(),
  Unmaximize: vi.fn(),
  Close: vi.fn(),
  ToggleDevTools: vi.fn(),
}));

describe("MenuBar Component", () => {
  const mockProps = {
    theme: "light",
    setTheme: vi.fn(),
    onOpenSettings: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders menu items", () => {
    render(<MenuBar {...mockProps} />);

    expect(screen.getByText("File")).toBeInTheDocument();
    expect(screen.getByText("View")).toBeInTheDocument();
    expect(screen.getByText("Settings")).toBeInTheDocument();
    expect(screen.getByText("Help")).toBeInTheDocument();
    expect(screen.getByText("Ostenia")).toBeInTheDocument();
  });

  it("opens and closes menus", () => {
    render(<MenuBar {...mockProps} />);

    const fileMenu = screen.getByText("File");
    fireEvent.click(fileMenu);

    expect(screen.getByText("Exit")).toBeInTheDocument();

    fireEvent.click(fileMenu);
    expect(screen.queryByText("Exit")).not.toBeInTheDocument();
  });

  it("handles window actions", () => {
    render(<MenuBar {...mockProps} />);

    const allButtons = screen.getAllByRole("button");
    const winCloseBtn = allButtons[allButtons.length - 1];
    const winMaxBtn = allButtons[allButtons.length - 2];
    const winMinBtn = allButtons[allButtons.length - 3];

    fireEvent.click(winMinBtn);
    expect(AppBackend.Minimize).toHaveBeenCalled();

    fireEvent.click(winMaxBtn);
    expect(AppBackend.Maximize).toHaveBeenCalled();

    fireEvent.click(winCloseBtn);
    expect(AppBackend.Close).toHaveBeenCalled();
  });

  it("handles settings category selection", () => {
    render(<MenuBar {...mockProps} />);

    fireEvent.click(screen.getByText("Settings"));
    fireEvent.click(screen.getByText("Profile"));

    expect(mockProps.onOpenSettings).toHaveBeenCalledWith("profile");
  });

  it("toggles developer tools", () => {
    render(<MenuBar {...mockProps} />);

    fireEvent.click(screen.getByText("View"));
    fireEvent.click(screen.getByText("Toggle Developer Tools"));

    expect(AppBackend.ToggleDevTools).toHaveBeenCalled();
  });
});
