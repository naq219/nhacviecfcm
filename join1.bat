@echo off
setlocal enabledelayedexpansion

set "output=user_repo_combined.txt"

(
    echo //------start file: user_repo.go -------
    type "internal\repository\pocketbase\user_repo.go"
    echo.
    echo //------end file: user_repo.go ------
    echo.
    echo //------start file: user_repo_test.go -------
    type "internal\repository\pocketbase\user_repo_test.go"
    echo.
    echo //------end file: user_repo_test.go ------
    echo.
    echo //------start file: reminder.go -------
    type "internal\models\reminder.go"
    echo.
    echo //------end file: reminder.go ------
    echo.
    echo //------start file: interface.go -------
    type "internal\repository\interface.go"
    echo.
    echo //------end file: interface.go ------
    echo.
    echo //------start file: db_utils.go -------
    type "internal\db\db_utils.go"
    echo.
    echo //------end file: db_utils.go ------
) > "%output%"

echo Done! Output saved to %output%
pause