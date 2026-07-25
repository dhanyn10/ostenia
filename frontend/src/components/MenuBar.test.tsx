import { render, screen, fireEvent, act } from "@testing-library/react";
import { vi, describe, it, expect, beforeEach } from "vitest";
import MenuBar from "./MenuBar";
import * as AppBackend from "../../wailsjs/go/backend/App";
import * as runtime from "../../wailsjs/runtime/runtime";

// Mock Wails backend functions
vi.mock("../../wailsjs/go/backend/App", () => ({
  Minimize: vi.fn(),
  Maximize: vi.fn(),
  Unmaximize: vi.fn(),
  Close: vi.fn(),
  ToggleDevTools: vi.fn(),
}));

// Mock Wails runtime functions
vi.mock("../../wailsjs/runtime/runtime", () => ({
  ScreenGetAll: vi.fn().mockResolvedValue([{ isCurrent: true, width: 1920, height: 1080 }]),
  WindowSetPosition: vi.fn(),
  WindowSetSize: vi.fn(),
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

  it("toggles maximize and unmaximize on subsequent clicks", () => {
    render(<MenuBar {...mockProps} />);

    const allButtons = screen.getAllByRole("button");
    const winMaxBtn = allButtons[allButtons.length - 2];

    // First click toggles maximize
    fireEvent.click(winMaxBtn);
    expect(AppBackend.Maximize).toHaveBeenCalled();
    expect(AppBackend.Unmaximize).not.toHaveBeenCalled();

    // Second click toggles unmaximize
    fireEvent.click(winMaxBtn);
    expect(AppBackend.Unmaximize).toHaveBeenCalled();
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

  it("opens snap menu on hover and allows snapping window", async () => {
    vi.useFakeTimers();
    render(<MenuBar {...mockProps} />);

    const container = screen.getByTestId("maximize-container");

    // Trigger hover
    fireEvent.mouseEnter(container);

    // Fast forward timers and flush React state updates
    act(() => {
      vi.advanceTimersByTime(450);
    });

    // Verify Snap Window header is displayed
    expect(screen.getByText("Snap Window")).toBeInTheDocument();

    // Find and click 'Left Half' snapping option
    const leftBtn = screen.getByText("Left Half").closest("button");
    expect(leftBtn).toBeInTheDocument();
    fireEvent.click(leftBtn!);

    // Run outstanding timers and promises
    await vi.runAllTimersAsync();

    // Verify that dynamic import called our mocked runtime
    expect(runtime.ScreenGetAll).toHaveBeenCalled();
    expect(runtime.WindowSetPosition).toHaveBeenCalledWith(0, 0);
    expect(runtime.WindowSetSize).toHaveBeenCalledWith(960, 1032); // 1920/2 and 1080-48

    vi.useRealTimers();
  });
});
