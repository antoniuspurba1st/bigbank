import pathlib
import re

auth_path = pathlib.Path(r'C:\Users\ronal\Desktop\com.bigbank\transaction-service\internal\handler\auth.go')
lines = auth_path.read_text(encoding='utf-8').splitlines(keepends=True)

output = []
i = 0
while i < len(lines):
    line = lines[i]
    
    # Fix 1: Replace /auth/me response to include user data
    if '"account_number": accountNumber,' in line and i+2 < len(lines) and '"transactions":   0,' in lines[i+2]:
        output.append('\t\t"id":             user.ID,\n')
        output.append('\t\t"email":          user.Email,\n')
        output.append('\t\t"phone":          user.Phone,\n')
        output.append('\t\t"account_number": accountNumber,\n')
        output.append('\t\t"balance":        balance,\n')
        i += 3  # skip old 3 lines
        continue
    
    # Fix 2: Replace non-standard error for account not found
    if 'writeJSON(w, http.StatusNotFound, map[string]interface{}{"error": "Account not found"})' in line:
        output.append('\t\t\twriteError(w, correlationIDFromRequest(r), &model.AppError{\n')
        output.append('\t\t\t\tStatusCode: http.StatusNotFound,\n')
        output.append('\t\t\t\tCode:       "ACCOUNT_NOT_FOUND",\n')
        output.append('\t\t\t\tMessage:    "Account not found",\n')
        output.append('\t\t\t})\n')
        i += 1
        continue
    
    # Fix 3: Remove internal error leak in register
    if '"Failed to create user: " + err.Error()' in line:
        output.append(line.replace('"Failed to create user: " + err.Error()', '"Failed to create user"'))
        i += 1
        continue
    
    output.append(line)
    i += 1

auth_path.write_text(''.join(output), encoding='utf-8')
print('auth.go fixed successfully')
