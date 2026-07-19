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
});
