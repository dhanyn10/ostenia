import { render, screen, fireEvent } from "@testing-library/react";
import { vi, describe, it, expect, beforeEach } from "vitest";
import MenuBar from "@/components/MenuBar";
import * as AppBackend from "@/wailsjs/go/backend/App";

// Mock Wails backend functions
vi.mock("@/wailsjs/go/backend/App", () => ({
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

  it("handles settings category selection and theme switching", () => {
    const { rerender } = render(<MenuBar {...mockProps} />);

    fireEvent.click(screen.getByText("Settings"));
    fireEvent.click(screen.getByText("Profile"));
    expect(mockProps.onOpenSettings).toHaveBeenCalledWith("profile");

    fireEvent.click(screen.getByText("Settings"));
    fireEvent.click(screen.getByText("Config"));
    expect(mockProps.onOpenSettings).toHaveBeenCalledWith("config");

    fireEvent.click(screen.getByText("Settings"));
    fireEvent.click(screen.getByText("SSH"));
    expect(mockProps.onOpenSettings).toHaveBeenCalledWith("ssh");

    fireEvent.click(screen.getByText("Settings"));
    fireEvent.click(screen.getByText("Switch to Dark Mode"));
    expect(mockProps.setTheme).toHaveBeenCalledWith("dark");

    rerender(<MenuBar {...mockProps} theme="dark" />);
    fireEvent.click(screen.getByText("Settings"));
    fireEvent.click(screen.getByText("Switch to Light Mode"));
    expect(mockProps.setTheme).toHaveBeenCalledWith("light");
  });

  it("handles help menu items including documentation and about dialog", () => {
    const windowOpenSpy = vi.spyOn(window, "open").mockImplementation(() => null);
    render(<MenuBar {...mockProps} />);

    fireEvent.click(screen.getByText("Help"));
    fireEvent.click(screen.getByText("Documentation"));
    expect(windowOpenSpy).toHaveBeenCalledWith(
      "https://github.com/dhanyn/ostenia",
      "_blank",
      "noopener,noreferrer",
    );

    fireEvent.click(screen.getByText("Help"));
    fireEvent.click(screen.getByText("About Ostenia"));
    expect(screen.getByText(/Ostenia v/i)).toBeInTheDocument();

    windowOpenSpy.mockRestore();
  });

  it("handles exit in file menu", () => {
    render(<MenuBar {...mockProps} />);

    fireEvent.click(screen.getByText("File"));
    fireEvent.click(screen.getByText("Exit"));
    expect(AppBackend.Close).toHaveBeenCalled();
  });

  it("toggles developer tools via click and keyboard", () => {
    render(<MenuBar {...mockProps} />);

    fireEvent.click(screen.getByText("View"));
    const devToolsBtn = screen.getByText("Toggle Developer Tools");
    fireEvent.click(devToolsBtn);
    expect(AppBackend.ToggleDevTools).toHaveBeenCalledTimes(1);

    fireEvent.click(screen.getByText("View"));
    const devToolsBtnAgain = screen.getByText("Toggle Developer Tools");
    fireEvent.keyDown(devToolsBtnAgain, { key: "Enter" });
    expect(AppBackend.ToggleDevTools).toHaveBeenCalledTimes(2);
  });
});
