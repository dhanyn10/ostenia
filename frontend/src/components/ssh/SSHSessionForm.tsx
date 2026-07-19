import React, { useState, useEffect } from "react";
import { X, Save, RefreshCw } from "lucide-react";
import { clsx } from "clsx";
import * as AppBackend from "../../../wailsjs/go/backend/App";

interface SSHSessionFormProps {
  session: any;
  onClose: () => void;
  onSave: () => void;
  addToast: (
    title: string,
    message: string,
    type?: "info" | "success" | "warn" | "error",
  ) => void;
}

const SSHSessionForm: React.FC<SSHSessionFormProps> = ({
  session,
  onClose,
  onSave,
  addToast,
}) => {
  const [formData, setFormData] = useState(
    session || {
      id: crypto.randomUUID(),
      name: "",
      host: "",
      port: 22,
      user: "root",
      authMethod: "password",
      password: "",
      keyPath: "",
      passphrase: "",
      createdAt: Date.now(),
      type: "ssh",
      wslDistro: "",
    },
  );

  const [wslDistros, setWslDistros] = useState<string[]>([]);
  const [loadingDistros, setLoadingDistros] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const fetchDistros = async () => {
      setLoadingDistros(true);
      try {
        const list = await AppBackend.GetWSLDistributions();
        setWslDistros(list || []);
        if (list && list.length > 0 && !formData.wslDistro) {
          setFormData((prev: any) => ({ ...prev, wslDistro: list[0] }));
        }
      } catch (err) {
        console.error("Failed to load WSL distros:", err);
      } finally {
        setLoadingDistros(false);
      }
    };
    fetchDistros();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      const dataToSave = { ...formData };
      if (dataToSave.type === "wsl") {
        dataToSave.host = dataToSave.wslDistro;
        dataToSave.user = "wsl";
        dataToSave.port = 0;
        dataToSave.authMethod = "wsl";
      }
      if (session) {
        await AppBackend.UpdateSSHSession(dataToSave);
      } else {
        await AppBackend.AddSSHSession(dataToSave);
      }
      addToast(
        "Success",
        `Session ${session ? "updated" : "created"} successfully`,
        "success",
      );
      onSave();
    } catch (err: any) {
      addToast("Error", "Failed to save session: " + err, "error");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="w-[380px] bg-white dark:bg-mui-dark-bg border-l border-mui-grey-200 dark:border-white/10 flex flex-col h-full animate-in slide-in-from-right duration-200 shrink-0">
      <div className="px-5 py-4 border-b border-mui-grey-100 dark:border-white/5 flex items-center justify-between bg-mui-grey-50 dark:bg-mui-grey-900/50">
        <div>
          <h3 className="text-sm font-bold text-mui-grey-900 dark:text-white uppercase tracking-wider">
            {session ? "Edit Connection" : "New Connection"}
          </h3>
        </div>
        <button
          type="button"
          onClick={onClose}
          className="p-1.5 hover:bg-mui-grey-200 dark:hover:bg-white/10 rounded text-mui-grey-500 transition-colors"
        >
          <X size={18} />
        </button>
      </div>

      <form
        onSubmit={handleSubmit}
        className="flex-1 overflow-y-auto p-5 space-y-6 custom-scrollbar"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-[10px] font-bold text-mui-grey-400 uppercase tracking-widest mb-1.5 ml-0.5">
              Connection Type
            </label>
            <div className="flex p-1 bg-mui-grey-100 dark:bg-mui-grey-900 rounded-md">
              <button
                type="button"
                onClick={() =>
                  setFormData({ ...formData, type: "ssh" })
                }
                className={clsx(
                  "flex-1 py-1.5 text-[11px] font-bold rounded transition-all",
                  formData.type !== "wsl"
                    ? "bg-white dark:bg-mui-blue-600 text-mui-blue-600 dark:text-white shadow-sm"
                    : "text-mui-grey-500 hover:text-mui-grey-700 dark:hover:text-mui-grey-300",
                )}
              >
                SSH
              </button>
              <button
                type="button"
                onClick={() => {
                  setFormData({ ...formData, type: "wsl" });
                  if (wslDistros.length > 0 && !formData.wslDistro) {
                    setFormData((prev: any) => ({ ...prev, wslDistro: wslDistros[0] }));
                  }
                }}
                className={clsx(
                  "flex-1 py-1.5 text-[11px] font-bold rounded transition-all",
                  formData.type === "wsl"
                    ? "bg-white dark:bg-mui-blue-600 text-mui-blue-600 dark:text-white shadow-sm"
                    : "text-mui-grey-500 hover:text-mui-grey-700 dark:hover:text-mui-grey-300",
                )}
              >
                WSL
              </button>
            </div>
          </div>

          <div>
            <label
              htmlFor="session-name"
              className="block text-[10px] font-bold text-mui-grey-400 uppercase tracking-widest mb-1.5 ml-0.5"
            >
              Label
            </label>
            <input
              id="session-name"
              required
              type="text"
              placeholder="e.g. Production Web"
              className="w-full px-3 py-2 bg-mui-grey-50 dark:bg-white/5 border border-transparent focus:border-mui-blue-500 focus:bg-white dark:focus:bg-mui-grey-900 rounded-md outline-none text-mui-grey-900 dark:text-white transition-all text-sm"
              value={formData.name}
              onChange={(e) =>
                setFormData({ ...formData, name: e.target.value })
              }
            />
          </div>

          {formData.type === "wsl" ? (
            <div>
              <label
                htmlFor="session-wslDistro"
                className="block text-[10px] font-bold text-mui-grey-400 uppercase tracking-widest mb-1.5 ml-0.5"
              >
                WSL Distribution
              </label>
              {loadingDistros ? (
                <div className="text-xs text-mui-grey-500 dark:text-mui-grey-400 py-2">
                  Loading distributions...
                </div>
              ) : wslDistros.length === 0 ? (
                <div className="text-xs text-red-500 dark:text-red-400 py-2">
                  No WSL distributions detected. Make sure WSL is installed and enabled.
                </div>
              ) : (
                <select
                  id="session-wslDistro"
                  required
                  className="w-full px-3 py-2 bg-mui-grey-50 dark:bg-white/5 border border-transparent focus:border-mui-blue-500 focus:bg-white dark:focus:bg-mui-grey-900 rounded-md outline-none text-mui-grey-900 dark:text-white transition-all text-sm"
                  value={formData.wslDistro}
                  onChange={(e) =>
                    setFormData({ ...formData, wslDistro: e.target.value })
                  }
                >
                  {wslDistros.map((distro) => (
                    <option key={distro} value={distro} className="text-black">
                      {distro}
                    </option>
                  ))}
                </select>
              )}
            </div>
          ) : (
            <>
              <div className="grid grid-cols-4 gap-2">
                <div className="col-span-3">
                  <label
                    htmlFor="session-host"
                    className="block text-[10px] font-bold text-mui-grey-400 uppercase tracking-widest mb-1.5 ml-0.5"
                  >
                    Address
                  </label>
                  <input
                    id="session-host"
                    required
                    type="text"
                    placeholder="1.2.3.4 or example.com"
                    className="w-full px-3 py-2 bg-mui-grey-50 dark:bg-white/5 border border-transparent focus:border-mui-blue-500 focus:bg-white dark:focus:bg-mui-grey-900 rounded-md outline-none text-mui-grey-900 dark:text-white transition-all text-sm"
                    value={formData.host}
                    onChange={(e) =>
                      setFormData({ ...formData, host: e.target.value })
                    }
                  />
                </div>
                <div className="col-span-1">
                  <label
                    htmlFor="session-port"
                    className="block text-[10px] font-bold text-mui-grey-400 uppercase tracking-widest mb-1.5 ml-0.5"
                  >
                    Port
                  </label>
                  <input
                    id="session-port"
                    required
                    type="number"
                    className="w-full px-2 py-2 bg-mui-grey-50 dark:bg-white/5 border border-transparent focus:border-mui-blue-500 focus:bg-white dark:focus:bg-mui-grey-900 rounded-md outline-none text-mui-grey-900 dark:text-white transition-all text-sm text-center"
                    value={formData.port}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        port: Number.parseInt(e.target.value),
                      })
                    }
                  />
                </div>
              </div>

              <div>
                <label
                  htmlFor="session-user"
                  className="block text-[10px] font-bold text-mui-grey-400 uppercase tracking-widest mb-1.5 ml-0.5"
                >
                  Username
                </label>
                <input
                  id="session-user"
                  required
                  type="text"
                  placeholder="root"
                  className="w-full px-3 py-2 bg-mui-grey-50 dark:bg-white/5 border border-transparent focus:border-mui-blue-500 focus:bg-white dark:focus:bg-mui-grey-900 rounded-md outline-none text-mui-grey-900 dark:text-white transition-all text-sm"
                  value={formData.user}
                  onChange={(e) =>
                    setFormData({ ...formData, user: e.target.value })
                  }
                />
              </div>

              <div>
                <label className="block text-[10px] font-bold text-mui-grey-400 uppercase tracking-widest mb-1.5 ml-0.5">
                  Authentication Method
                </label>
                <div className="flex p-1 bg-mui-grey-100 dark:bg-mui-grey-900 rounded-md">
                  <button
                    type="button"
                    onClick={() =>
                      setFormData({ ...formData, authMethod: "password" })
                    }
                    className={clsx(
                      "flex-1 py-1.5 text-[11px] font-bold rounded transition-all",
                      formData.authMethod === "password"
                        ? "bg-white dark:bg-mui-blue-600 text-mui-blue-600 dark:text-white shadow-sm"
                        : "text-mui-grey-500 hover:text-mui-grey-700 dark:hover:text-mui-grey-300",
                    )}
                  >
                    Password
                  </button>
                  <button
                    type="button"
                    onClick={() => setFormData({ ...formData, authMethod: "key" })}
                    className={clsx(
                      "flex-1 py-1.5 text-[11px] font-bold rounded transition-all",
                      formData.authMethod === "key"
                        ? "bg-white dark:bg-mui-blue-600 text-mui-blue-600 dark:text-white shadow-sm"
                        : "text-mui-grey-500 hover:text-mui-grey-700 dark:hover:text-mui-grey-300",
                    )}
                  >
                    Key File
                  </button>
                </div>
              </div>

              {formData.authMethod === "password" ? (
                <div>
                  <label
                    htmlFor="session-password"
                    className="block text-[10px] font-bold text-mui-grey-400 uppercase tracking-widest mb-1.5 ml-0.5"
                  >
                    Password
                  </label>
                  <input
                    id="session-password"
                    required
                    type="password"
                    placeholder="••••••••"
                    className="w-full px-3 py-2 bg-mui-grey-50 dark:bg-white/5 border border-transparent focus:border-mui-blue-500 focus:bg-white dark:focus:bg-mui-grey-900 rounded-md outline-none text-mui-grey-900 dark:text-white transition-all text-sm"
                    value={formData.password}
                    onChange={(e) =>
                      setFormData({ ...formData, password: e.target.value })
                    }
                  />
                </div>
              ) : (
                <div className="space-y-4">
                  <div>
                    <label
                      htmlFor="session-keyPath"
                      className="block text-[10px] font-bold text-mui-grey-400 uppercase tracking-widest mb-1.5 ml-0.5"
                    >
                      Private Key Path
                    </label>
                    <input
                      id="session-keyPath"
                      required
                      type="text"
                      placeholder="e.g. /home/user/.ssh/id_rsa"
                      className="w-full px-3 py-2 bg-mui-grey-50 dark:bg-white/5 border border-transparent focus:border-mui-blue-500 focus:bg-white dark:focus:bg-mui-grey-900 rounded-md outline-none text-mui-grey-900 dark:text-white transition-all text-sm"
                      value={formData.keyPath}
                      onChange={(e) =>
                        setFormData({ ...formData, keyPath: e.target.value })
                      }
                    />
                  </div>
                  <div>
                    <label
                      htmlFor="session-passphrase"
                      className="block text-[10px] font-bold text-mui-grey-400 uppercase tracking-widest mb-1.5 ml-0.5"
                    >
                      Passphrase
                    </label>
                    <input
                      id="session-passphrase"
                      type="password"
                      placeholder="optional"
                      className="w-full px-3 py-2 bg-mui-grey-50 dark:bg-white/5 border border-transparent focus:border-mui-blue-500 focus:bg-white dark:focus:bg-mui-grey-900 rounded-md outline-none text-mui-grey-900 dark:text-white transition-all text-sm"
                      value={formData.passphrase}
                      onChange={(e) =>
                        setFormData({ ...formData, passphrase: e.target.value })
                      }
                    />
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </form>

      <div className="p-4 border-t border-mui-grey-100 dark:border-white/5 bg-white dark:bg-mui-dark-bg flex gap-2">
        <button
          type="button"
          onClick={onClose}
          className="flex-1 py-2 text-xs font-bold text-mui-grey-500 hover:text-mui-grey-900 dark:hover:text-white transition-colors border border-mui-grey-200 dark:border-white/10 rounded-md"
        >
          Cancel
        </button>
        <button
          type="submit"
          onClick={handleSubmit}
          disabled={saving}
          className="flex-1 py-2 bg-mui-blue-600 hover:bg-mui-blue-700 text-white text-xs font-bold rounded-md flex items-center justify-center gap-2 transition-all disabled:opacity-50"
        >
          {saving ? (
            <RefreshCw className="animate-spin" size={14} />
          ) : (
            <Save size={14} />
          )}
          {session ? "Update" : "Save"}
        </button>
      </div>
    </div>
  );
};

export default SSHSessionForm;
