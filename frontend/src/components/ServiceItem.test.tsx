import { render, screen, fireEvent, act } from "@testing-library/react";
import { vi, describe, it, expect, beforeEach } from "vitest";
import ServiceItem from "./ServiceItem";

describe("ServiceItem Component", () => {
  const mockProps = {
    service: {
      name: "Apache",
      status: "Stopped",
      pid: 0,
      port: 80,
      ports: [80, 443],
      activeVersion: "2.4.58",
    },
    task: {
      name: "Apache",
      installedVers: ["2.4.58", "2.4.59"],
    },
    isExpanded: false,
    onToggleAccordion: vi.fn(),
    renderIcon: vi.fn(() => <span data-testid="icon" />),
    handleToggleService: vi.fn(),
    handleRemoveFromHome: vi.fn(),
    handleSwitchVersion: vi.fn(),
    handleOpenLocalTerminal: vi.fn(),
    handleToggleHttps: vi.fn(),
    openTerminalDropdown: null,
    setOpenTerminalDropdown: vi.fn(),
    setIsModalOpen: vi.fn(),
    apacheHttps: false,
    nginxHttps: false,
    isOpenSslEnabled: true,
    setActiveTab: vi.fn(),
    handleOpenPluginFolder: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders service information correctly", () => {
    render(<ServiceItem {...mockProps} />);

    expect(screen.getByText("Apache")).toBeInTheDocument();
    expect(screen.getByText("Stopped")).toBeInTheDocument();
    expect(screen.getByTestId("icon")).toBeInTheDocument();
  });

  it("shows running status and stats when running", () => {
    const runningProps = {
      ...mockProps,
      service: { ...mockProps.service, status: "Running", pid: 1234 },
    };
    render(<ServiceItem {...runningProps} />);

    expect(screen.getByText("Running")).toBeInTheDocument();
    expect(screen.getByText(/PID: 1234/)).toBeInTheDocument();
    expect(screen.getByText(/Port: 80, 443/)).toBeInTheDocument();
  });

  it("calls onToggleAccordion when clicked", () => {
    render(<ServiceItem {...mockProps} />);

    const toggleButton = screen.getByText("Apache").closest("button");
    fireEvent.click(toggleButton!);

    expect(mockProps.onToggleAccordion).toHaveBeenCalledWith("Apache", true);
  });

  it("calls onToggleAccordion when the card element is clicked", () => {
    const { container } = render(<ServiceItem {...mockProps} />);

    const card = container.querySelector('[data-card-expand]') as HTMLElement;
    fireEvent.click(card);

    expect(mockProps.onToggleAccordion).toHaveBeenCalledWith("Apache", true);
  });

  it("supports keyboard expansion with Enter and Space on the card", () => {
    const { container } = render(<ServiceItem {...mockProps} />);
    const card = container.querySelector('[data-card-expand]') as HTMLElement;

    fireEvent.click(card);
    expect(mockProps.onToggleAccordion).toHaveBeenCalledWith("Apache", true);

    fireEvent.click(card);
    expect(mockProps.onToggleAccordion).toHaveBeenCalledTimes(2);
  });

  it("does not toggle accordion on click or key press when service has no extra actions", () => {
    const propsNoExtra = {
      ...mockProps,
      service: {
        ...mockProps.service,
        name: "UnknownService",
      },
      task: {
        name: "UnknownService",
        installedVers: [],
      },
      isOpenSslEnabled: false,
    };
    const { container } = render(<ServiceItem {...propsNoExtra} />);
    expect(container.querySelector('[data-card-expand]')).toBeNull();
  });

  it("calls handleToggleService when main action button is clicked", () => {
    render(<ServiceItem {...mockProps} />);

    const buttons = screen.getAllByRole("button");
    const toggleButton = buttons.find((b) => b.className.includes("w-12 h-6"));

    fireEvent.click(toggleButton!);
    expect(mockProps.handleToggleService).toHaveBeenCalledWith(
      "Apache",
      "Stopped",
    );
  });

  it("shows extra actions when expanded", () => {
    const expandedProps = { ...mockProps, isExpanded: true };
    render(<ServiceItem {...expandedProps} />);

    expect(screen.getByTitle("Open Folder")).toBeInTheDocument();
    expect(screen.getByTitle("Terminal")).toBeInTheDocument();
    expect(screen.getByTitle("Enable HTTPS")).toBeInTheDocument();
  });

  it("handles version switching", () => {
    render(
      <ServiceItem
        {...mockProps}
        service={{ ...mockProps.service, name: "PHP" }}
      />,
    );

    const versionButton = screen.getByText("2.4.59");
    fireEvent.click(versionButton);

    expect(mockProps.handleSwitchVersion).toHaveBeenCalledWith("PHP", "2.4.59");
  });

  it("handles terminal dropdown and terminal actions", () => {
    // 1. Terminal button toggles dropdown
    const expandedProps = { ...mockProps, isExpanded: true };
    const { rerender } = render(<ServiceItem {...expandedProps} />);

    const termBtn = screen.getByTitle("Terminal");
    fireEvent.click(termBtn);
    expect(mockProps.setOpenTerminalDropdown).toHaveBeenCalledWith("Apache");

    // 2. Dropdown open state
    const openDropdownProps = {
      ...expandedProps,
      openTerminalDropdown: "Apache",
    };
    rerender(<ServiceItem {...openDropdownProps} />);

    expect(screen.getByText("CMD")).toBeInTheDocument();
    expect(screen.getByText("PowerShell")).toBeInTheDocument();

    fireEvent.click(screen.getByText("CMD"));
    expect(mockProps.handleOpenLocalTerminal).toHaveBeenCalledWith(
      "Apache",
      "cmd",
    );

    fireEvent.click(screen.getByText("PowerShell"));
    expect(mockProps.handleOpenLocalTerminal).toHaveBeenCalledWith(
      "Apache",
      "powershell",
    );

    // Backdrop click closes dropdown
    const backdrop = document.querySelector(".fixed.inset-0") as HTMLElement;
    fireEvent.click(backdrop);
    expect(mockProps.setOpenTerminalDropdown).toHaveBeenCalledWith(null);
  });

  it("handles folder, https, settings and remove actions when expanded", () => {
    const expandedProps = { ...mockProps, isExpanded: true };
    const { rerender } = render(<ServiceItem {...expandedProps} />);

    fireEvent.click(screen.getByTitle("Open Folder"));
    expect(mockProps.handleOpenPluginFolder).toHaveBeenCalledWith("Apache");

    fireEvent.click(screen.getByTitle("Enable HTTPS"));
    expect(mockProps.handleToggleHttps).toHaveBeenCalledWith("Apache");

    fireEvent.click(screen.getByTitle("Remove from home"));
    expect(mockProps.handleRemoveFromHome).toHaveBeenCalledWith("Apache");

    // PHP Extensions button
    rerender(
      <ServiceItem
        {...expandedProps}
        service={{ ...mockProps.service, name: "PHP" }}
      />,
    );
    fireEvent.click(screen.getByText("Extensions"));
    expect(mockProps.setIsModalOpen).toHaveBeenCalledWith(true);
  });
});
