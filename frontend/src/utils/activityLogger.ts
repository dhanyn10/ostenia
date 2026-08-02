/**
 * Measures the execution time of a synchronous or asynchronous action
 * and logs the performance and results to the console.
 *
 * Since console functions are overridden, these logs are automatically
 * captured by the stack trace parser.
 *
 * @param funcName The name of the function/activity to log.
 * @param action The synchronous or asynchronous function to execute.
 * @returns The result of the action.
 */
export async function measureActivity<T>(
  funcName: string,
  action: () => Promise<T> | T
): Promise<T> {
  const startTime = performance.now();
  try {
    const result = await action();
    const duration = (performance.now() - startTime).toFixed(1);
    console.log(`Activity: ${funcName} executed in ${duration}ms. Result: SUCCESS`);
    return result;
  } catch (err: any) {
    const duration = (performance.now() - startTime).toFixed(1);
    const errMsg = err?.message || String(err);
    console.error(`Activity: ${funcName} executed in ${duration}ms. Result: ERROR - ${errMsg}`);
    throw err;
  }
}
