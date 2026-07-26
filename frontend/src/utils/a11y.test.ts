import { handleActionKey } from "./a11y";
import { describe, it, expect, vi } from "vitest";

describe("a11y utilities", () => {
  it("triggers callback on Enter and Space key presses", () => {
    const callback = vi.fn();
    const handler = handleActionKey(callback);

    // Mock KeyboardEvents
    const mockPreventDefault = vi.fn();

    // 1. Enter key
    const enterEvent = {
      key: "Enter",
      preventDefault: mockPreventDefault,
    } as any;
    handler(enterEvent);
    expect(callback).toHaveBeenCalledTimes(1);
    expect(mockPreventDefault).toHaveBeenCalledTimes(1);

    // 2. Space key
    const spaceEvent = {
      key: " ",
      preventDefault: mockPreventDefault,
    } as any;
    handler(spaceEvent);
    expect(callback).toHaveBeenCalledTimes(2);
    expect(mockPreventDefault).toHaveBeenCalledTimes(2);

    // 3. Other key (e.g. Escape)
    const otherEvent = {
      key: "Escape",
      preventDefault: mockPreventDefault,
    } as any;
    handler(otherEvent);
    expect(callback).toHaveBeenCalledTimes(2); // no increase
  });
});
