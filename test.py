import ctypes
import ctypes.wintypes
import string

def generate_keyboard_mapping():
    mapping = {}
    for char in string.printable:  # all printable ASCII characters
        vk_code = ctypes.windll.user32.VkKeyScanW(ctypes.wintypes.WCHAR(char))
        if vk_code != -1:
            mapping[char] = vk_code
    return mapping

keyboardMapping = generate_keyboard_mapping()

# Print the mapping in a readable way
for char, code in keyboardMapping.items():
    mods, vk = divmod(code, 0x100)
    print(f"'{char}': VK={vk}, mods={mods}")
