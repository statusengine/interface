<?php
/**
 * Statusengine UI
 * Copyright (C) 2019  Daniel Ziegler
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine;


class Env {

    const VALUE_STRING = 1;
    const VALUE_ARRAY  = 2;
    const VALUE_BOOL   = 4;
    const VALUE_INT    = 8;

    /**
     * @param string $varName   Name of the environment variable (must start with SE_)
     * @param mixed $default    Default value which will be returned if the env variable is not set
     * @param int|null $type    Type of the variable value (must be null for self::VALUE_STRING, self::VALUE_ARRAY, or self::VALUE_BOOL)
     * @return mixed            Date read from environment variable or passed default value
     */
    public static function get($varName, $default, $type = null) {
        if ($type === null) {
            $type = self::VALUE_STRING;
        }

        if (isset($_SERVER[$varName]) && strpos($varName, 'SEI_', 0) === 0) {
            $value = $_SERVER[$varName];

            if ($type === self::VALUE_BOOL) {
                $value = strtolower($value);
                if ($value === 'true' || $value === 'on' || $value === '1' || $value === 1) {
                    return true;
                }

                return false;
            }

            if ($type === self::VALUE_ARRAY) {
                return explode(',', $value);
            }

            if ($type === self::VALUE_INT) {
                return (int)$value;
            }

            return $value;
        }
        return $default;
    }

}
